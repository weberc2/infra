package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"

	"github.com/weberc2/infra/scripts/generate-workflows/pkg/projects"
)

func findRepoRoot(dir string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		if os.IsNotExist(err) {
			return findRepoRoot(filepath.Dir(dir))
		}
		return "", err
	}

	return dir, nil
}

func main() {
	if err := entrypoint(); err != nil {
		chunks := strings.Split(err.Error(), ": ")
		indent := ""
		for _, chunk := range chunks {
			color.Red("%s↪️ ️%s\n", indent, chunk)
			indent += "  "
		}
		os.Exit(1)
	}
}

func entrypoint() error {
	// Create a temporary directory to represent the final
	// `~/.github/workflows` directory. If all goes well, we'll do a rename at
	// the end to atomically "promote" this temporary directory to become the
	// official `~/.github/workflows` directory.
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("Creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Find the root of the repository
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Getting working directory: %w", err)
	}
	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		return fmt.Errorf("Finding repo root: %w", err)
	}

	dir := filepath.Join(repoRoot, ".github/workflows")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Build and render project workflow files
	if err := projects.RenderProjectWorkflows(projectTypes, repoRoot, tmpDir); err != nil {
		return fmt.Errorf("Rendering project workflows: %w", err)
	}
	success("Staged project workflows")

	// Render the static files
	for fileName, contents := range staticFiles {
		filePath := filepath.Join(tmpDir, fileName)
		if err := func() error {
			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("Creating static file '%s': %w", filePath, err)
			}
			defer file.Close()

			if _, err := file.WriteString(contents); err != nil {
				return fmt.Errorf("Writing to static file '%s': %w", filePath, err)
			}
			return nil
		}(); err != nil {
			return err
		}
		success("Staged %s", fileName)
	}

	// Atomically "commit" the changes to `~/.github/workflows`.
	if err := os.Rename(tmpDir, dir); err != nil {
		if os.IsExist(err) {
			if err := os.RemoveAll(dir); err != nil {
				return fmt.Errorf("Removing dir '%s': %w", dir, err)
			}
			if err := os.Rename(tmpDir, dir); err != nil {
				return fmt.Errorf("Renaming '%s' to '%s': %w", tmpDir, dir, err)
			}
		} else {
			return fmt.Errorf("Renaming '%s' to '%s': %w", tmpDir, dir, err)
		}
	}
	success("Promoted staged files")
	return nil
}

var golangProjectType = projects.ProjectType{
	Identifier: "golang",
	Workflows: projects.WorkflowTypes{
		projects.WorkflowPullRequest: {golangTestJobType, golangLintJobType},
		projects.WorkflowMerge:       {golangTestJobType, golangLintJobType},
	},
}

var projectTypes = []projects.ProjectType{
	{
		Identifier: "golanglambda",
		Dependencies: map[string]*projects.ProjectType{
			"golang-source-project": &golangProjectType,
		},
		Workflows: projects.WorkflowTypes{
			projects.WorkflowPullRequest: {
				{
					Name: "greet",
					Dependencies: []projects.JobTypeDependency{{
						Name:     "golang-source-project",
						JobIndex: 0, // test
					}, {
						Name:     "golang-source-project",
						JobIndex: 1,
					}},
					RunsOn: "ubuntu-latest",
					Steps: []projects.JobStep{
						{Uses: "actions/checkout@v2"},
						{Name: "Do something", Run: "echo \"Hello, world!\""},
					},
				},
			},
			projects.WorkflowMerge: {
				{
					Name: "s3publish",
					Dependencies: []projects.JobTypeDependency{{
						Name:     "golang-source-project",
						JobIndex: 0, // test
					}, {
						Name:     "golang-source-project",
						JobIndex: 1,
					}},
					RunsOn: "ubuntu-latest",
					Steps: []projects.JobStep{
						{Uses: "actions/checkout@v2"},
						{Uses: "actions/setup-go@v2"},
						{
							Name: "Build binary",
							Run: `set -eo pipefail
cd {{ .Path }}
output="$PWD/{{ .Name }}-$(git rev-parse HEAD)"
echo "output=$output" >> $GITHUB_ENV
go build -o "$output"`,
						},
						{
							Name: "Zip artifact",
							Run:  "echo \"$output\" && zip \"${output}.zip\" \"$output\"",
						},
						{
							Name: "Publish to S3",
							Env: map[string]string{
								"AWS_ACCESS_KEY_ID":     "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}",
								"AWS_SECRET_ACCESS_KEY": "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}",
								"AWS_DEFAULT_REGION":    "us-east-2",
							},
							Run: `aws --debug s3 cp "${output}.zip" "s3://weberc2-inf-lambda-code-artifacts/$(basename $output).zip"`,
						},
					},
				},
			},
		},
	},
	golangProjectType,
	{
		Identifier: "terraformtarget",
		Workflows: projects.WorkflowTypes{
			projects.WorkflowPullRequest: {
				{
					Name:   "plan",
					RunsOn: "ubuntu-latest",
					Steps: []projects.JobStep{
						{Uses: "actions/checkout@v2"},
						{Name: "Terraform setup", Uses: "hashicorp/setup-terraform@v1"},
						{
							Name: "Terraform init",
							Env: map[string]string{
								"AWS_ACCESS_KEY_ID":     "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}",
								"AWS_SECRET_ACCESS_KEY": "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}",
							},
							Run: "terraform -chdir={{ .Path }} init",
						},
						{
							Name: "Terraform plan",
							Env: map[string]string{
								"AWS_ACCESS_KEY_ID":     "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}",
								"AWS_SECRET_ACCESS_KEY": "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}",
							},
							Run: "terraform -chdir={{ .Path }} plan",
						},
					},
				},
			},
			projects.WorkflowMerge: {
				{
					Name:   "apply",
					RunsOn: "ubuntu-latest",
					Steps: []projects.JobStep{
						{Uses: "actions/checkout@v2"},
						{Name: "Terraform setup", Uses: "hashicorp/setup-terraform@v1"},
						{
							Name: "Terraform init",
							Env: map[string]string{
								"AWS_ACCESS_KEY_ID":     "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}",
								"AWS_SECRET_ACCESS_KEY": "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}",
							},
							Run: "terraform -chdir={{ .Path }} init",
						},
						{
							Name: "Terraform apply",
							Env: map[string]string{
								"AWS_ACCESS_KEY_ID":     "${{ secrets.TERRAFORM_AWS_ACCESS_KEY_ID }}",
								"AWS_SECRET_ACCESS_KEY": "${{ secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY }}",
							},
							Run: "terraform -chdir={{ .Path }} apply",
						},
					},
				},
			},
		},
	},
}

var golangLintJobType = projects.JobType{
	Name:   "lint",
	RunsOn: "ubuntu-latest",
	Steps: []projects.JobStep{
		{Uses: "actions/checkout@v2"},
		{Uses: "actions/setup-go@v2"},
		{
			Name: "Fetch golint",
			Run: `export GOBIN=$PWD/{{ .Path }}/bin
echo "GOBIN=$GOBIN" >> $GITHUB_ENV
(cd {{ .Path }} && go get golang.org/x/lint/golint)
`,
		},
		{
			Name: "Lint",
			Run:  "(cd {{ .Path }} && $GOBIN/golint -set_exit_status ./...)",
		},
	},
}

var golangTestJobType = projects.JobType{
	Name:   "test",
	RunsOn: "ubuntu-latest",
	Steps: []projects.JobStep{
		{Uses: "actions/checkout@v2"},
		{Uses: "actions/setup-go@v2"},
		{Name: "Test", Run: "(cd {{ .Path }} && go test -v ./...)"},
	},
}

var staticFiles = map[string]string{
	"terraform-fmt.yaml": `name: Terraform format check

on:
  pull_request:
    branches: [ master ]

jobs:
  format-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Terraform setup
        uses: hashicorp/setup-terraform@v1
      - name: Terraform format check
        run: terraform fmt -recursive -check
`,
	"generate-workflows-check.yaml": `name: Generate workflows check

on:
  pull_request:
    branches: [ master ]

jobs:
  generate-workflows-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - name: Run script
        run: (cd scripts/generate-workflows && go run .)
      - name: Check diff
        run: |
          if [[ -n "$(git diff .github/workflows)" ]]; then
              echo "Unexpected differences in the .github/workflows directory:"
              git diff .github/workflows
              echo ""
              echo "Run ./scripts/generate-workflows.go from the repo root and commit the results."
              exit 1
          fi
`,
}

func success(format string, v ...interface{}) {
	fmt.Printf("✅ "+format+"\n", v...)
}
