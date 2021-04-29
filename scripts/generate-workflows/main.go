package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
	// Create a temporary directory to represent the final
	// `~/.github/workflows` directory. If all goes well, we'll do a rename at
	// the end to atomically "promote" this temporary directory to become the
	// official `~/.github/workflows` directory.
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		fatal("Creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Find the root of the repository
	cwd, err := os.Getwd()
	if err != nil {
		fatal("Getting working directory: %v", err)
	}
	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		fatal("Finding repo root: %v", err)
	}

	dir := filepath.Join(repoRoot, ".github/workflows")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Build and render project workflow files
	if err := projects.RenderProjectWorkflows(projectTypes, repoRoot, tmpDir); err != nil {
		fatal("Render project workflows: %v", err)
	}
	success("Staged project workflows")

	// Render the static files
	for fileName, contents := range staticFiles {
		filePath := filepath.Join(tmpDir, fileName)
		func() {
			file, err := os.Create(filePath)
			if err != nil {
				fatal("Creating static file '%s': %v", filePath, err)
			}
			defer file.Close()

			if _, err := file.WriteString(contents); err != nil {
				fatal("Writing to static file '%s': %v", filePath, err)
			}
		}()
		success("Staged %s", fileName)
	}

	// Atomically "commit" the changes to `~/.github/workflows`.
	if err := os.Rename(tmpDir, dir); err != nil {
		if os.IsExist(err) {
			if err := os.RemoveAll(dir); err != nil {
				fatal("Removing dir '%s': %v", dir, err)
			}
			if err := os.Rename(tmpDir, dir); err != nil {
				fatal("Renaming '%s' to '%s': %v", tmpDir, dir, err)
			}
		} else {
			fatal("Renaming '%s' to '%s': %v", tmpDir, dir, err)
		}
	}

	success("Promoted staged files")
}

func makeTemplate(name, body string) *template.Template {
	return template.Must(
		template.New(name).Parse(strings.Replace(body, "\t", "    ", -1)),
	)
}

var projectTypes = []projects.ProjectType{
	{
		Identifier: "golang",
		KeyFile:    "go.mod",
		Workflows: [projects.WorkflowMax][]projects.JobType{
			projects.WorkflowPullRequest: {golangTestJobType, golangLintJobType},
			projects.WorkflowMerge:       {golangTestJobType, golangLintJobType},
		},
	},
	{
		Identifier: "terraformtarget",
		KeyFile:    "terraform.tf",
		Workflows: [projects.WorkflowMax][]projects.JobType{
			projects.WorkflowPullRequest: {
				{
					Name: "plan",
					Template: makeTemplate(
						"plan",
						`{{ .Name }}-plan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - name: Terraform setup
      uses: hashicorp/setup-terraform@v1
    - name: Terraform init
      env:
        AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
        AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
      run: terraform -chdir={{ .Path }} init
    - name: Terraform plan
      env:
        AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
        AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
      run: terraform -chdir={{ .Path }} plan
`,
					),
				},
			},
			projects.WorkflowMerge: {
				{
					Name: "apply",
					Template: makeTemplate(
						"apply",
						`{{ .Name }}-apply:
  runs-on: ubuntu-latest
  steps:
    - name: Checkout
      uses: actions/checkout@v2
    - name: Terraform setup
      uses: hashicorp/setup-terraform@v1
    - name: Terraform init {{ .Name }}
      id: init-{{ .Name }}
      env:
        AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
        AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
      run: terraform -chdir={{ .Path }} init
    - name: Terraform apply {{ .Name }}
      id: apply-{{ .Name }}
      env:
        AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
        AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
      run: terraform -chdir={{ .Path }} apply -auto-approve
`,
					),
				},
			},
		},
	},
}

var golangLintJobType = projects.JobType{
	Name: "lint",
	Template: makeTemplate(
		"lint",
		`{{ .Name }}-lint:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
	- name: Fetch golint
	  run: |
	    export GOBIN=$PWD/{{ .Path }}/bin
		echo "GOBIN=$GOBIN" >> $GITHUB_ENV
	    (cd {{ .Path }} && go get golang.org/x/lint/golint)
    - name: Lint
      # Evidently we can't 'go test {{ .Path }}/...' or the go tool will
      # search GOPATH instead of the module at {{ .Path }}.
      run: (cd {{ .Path }} && $GOBIN/golint -set_exit_status .)
`,
	),
}

var golangTestJobType = projects.JobType{
	Name: "test",
	Template: makeTemplate(
		"test",
		`{{ .Name }}-test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
    - name: Test
      # Evidently we can't 'go test {{ .Path }}/...' or the go tool will
      # search GOPATH instead of the module at {{ .Path }}.
      run: cd {{ .Path }} && go test -v ./...
`,
	),
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
	log.Println(color.GreenString("✅ "+format, v...))
}

func fatal(format string, v ...interface{}) {
	log.Fatal(color.RedString("❌ "+format, v...))
}
