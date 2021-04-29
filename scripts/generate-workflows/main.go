package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

// WorkflowIdentifier is an enum whose variants identify different workflows.
type WorkflowIdentifier int

const (
	// WorkflowPullRequest identifies the PullRequest workflow.
	WorkflowPullRequest WorkflowIdentifier = iota

	// WorkflowMerge identifies the Merge workflow
	WorkflowMerge

	// WorkflowMax is the 'length' of the valid workflow identifiers.  It's not
	// a valid WorkflowIdentifier itself, but rather it's used for arrays which
	// are indexed by WorkflowIdentifiers to designate the length.  E.g.,
	// `[WorkflowMax][]Job`.
	WorkflowMax
)

// String returns the human-readable string representation of a WorkflowIdentifier.
func (wid WorkflowIdentifier) String() string {
	switch wid {
	case WorkflowPullRequest:
		return "Pull Request"
	case WorkflowMerge:
		return "Merge"
	default:
		panic(fmt.Sprintf("Invalid WorkflowIdentifier: %d", wid))
	}
}

// Trigger returns the GitHub Actions trigger key for the WorkflowIdentifier.
// (E.g., the `pull_request` bit in `on: pull_request: ...`).
func (wid WorkflowIdentifier) Trigger() string {
	switch wid {
	case WorkflowPullRequest:
		return "pull_request"
	case WorkflowMerge:
		return "push"
	default:
		panic(fmt.Sprintf("Invalid WorkflowIdentifier: %d", wid))
	}
}

// FileName returns the workflow filename that corresponds to the
// WorkflowIdentifier.
func (wid WorkflowIdentifier) FileName() string {
	switch wid {
	case WorkflowPullRequest:
		return "pull-request.yaml"
	case WorkflowMerge:
		return "merge.yaml"
	default:
		panic(fmt.Sprintf("Invalid WorkflowIdentifier: %d", wid))
	}
}

// JobType is the template or prototype from which `Job`s are created.  It
// associates a job name with a text template.
type JobType struct {
	// Name is suffixed onto all jobs of this job type.
	Name string

	// Template is the templates which will be rendered for
	// a given project of this project type. The resulting job name will be
	// structured `${project.Name()}-${job.Name}`, and it will be rendered onto
	// the file which corresponds to the job's workflow.
	Template *template.Template
}

// ProjectType represents a kind of project, e.g., a Go project, a Terraform
// project, a lambda project, etc. Each type of project is associated with a
// "key file" or a file in the root of the project which identifies the type of
// the file (see the `KeyFile` field for more information) as well as a
// collection of text templates which will be rendered for a project of this
// type into the `~/.github/workflows` output directory.  See the `Templates`
// field for more information.
type ProjectType struct {
	// Identifier will be prepended onto project names to disambiguate between
	// projects with the same name but different project types.
	Identifier string

	// KeyFile is the file to look for which will identify a directory as a
	// project of this ProjectType. E.g., for a Go project it will be `go.mod`,
	// for a Terraform project it will be `terraform.tf`, etc. It is possible
	// though probably inadvisable for a given directory to yield projects of
	// multiple types.
	KeyFile string

	// Workflows holds the `JobType`s associated with this project organized by
	// the workflow for which they're intended.  Namely, the key for the array
	// is intended to be a `WorkflowIdentifier` whose values are less than
	// `WorkflowMax`.
	Workflows [WorkflowMax][]JobType
}

// Project represents a project in a repository. Each project has a type (see
// `ProjectType` for more information) and a path (relative to the repo root).
type Project struct {
	// Type is the type of a project. It is used to render the final output
	// files associated with this project into the `~/.github/workflows`
	// directory.
	Type *ProjectType

	// Path is the path to the project directory relative to the root of the
	// repository. The "name" of the project is the basename of the path
	// prefixed by the project's type identifier (see
	// `ProjectType.Identifier`).
	Path string
}

// Name returns the name of the project by appending the basename of the
// project's `Path` with the project's type's identifier. This will be used as
// a parameter to template the output files.
func (p *Project) Name() string {
	return fmt.Sprintf("%s-%s", p.Type.Identifier, filepath.Base(p.Path))
}

// FindProjects searches the repo root to locate project directories and builds
// `Project`s from them. It will return an error if multiple projects were
// detected with the same basename and type.
func FindProjects(types []ProjectType, repoRoot string) ([]Project, error) {
	projects, err := findProjects(types, repoRoot, repoRoot)
	if err != nil {
		return nil, err
	}

	sort.Slice(projects, func(i, j int) bool {
		pi, pj := projects[i], projects[j]
		return pi.Type.Identifier < pj.Type.Identifier && pi.Name() < pj.Name()
	})

	for i := range projects[:len(projects)-1] {
		pi, pj := projects[i], projects[i+1]
		if pi.Type.Identifier == pj.Type.Identifier && pi.Name() == pj.Name() {
			return nil, fmt.Errorf(
				"duplicate projects detected: '%s' and '%s': two projects "+
					"may not share the same basename and project type",
				pi.Path,
				pj.Path,
			)
		}
	}

	return projects, nil
}

// findProjects is a recursive helper for `FindProjects`.
func findProjects(types []ProjectType, root, dir string) ([]Project, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var projects []Project
	for _, file := range files {
		if file.IsDir() {
			ps, err := findProjects(
				types,
				root,
				filepath.Join(dir, file.Name()),
			)
			if err != nil {
				return nil, err
			}

			projects = append(projects, ps...)
			continue
		}

		for i, projectType := range types {
			if file.Name() == projectType.KeyFile {
				path, err := filepath.Rel(root, dir)
				if err != nil {
					return nil, err
				}

				projects = append(
					projects,
					Project{
						Type: &types[i],
						Path: path,
					},
				)
			}
		}
	}

	return projects, nil
}

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

	// Collect the projects from the repository
	projects, err := FindProjects(projectTypes, repoRoot)
	if err != nil {
		fatal("Collecting projects: %v", err)
	}

	// Assemble and render the workflow files
	if err := Render(tmpDir, MaterializeWorkflows(projects)); err != nil {
		fatal("rendering workflows: %v", err)
	}

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

var projectTypes = []ProjectType{
	{
		Identifier: "golang",
		KeyFile:    "go.mod",
		Workflows: [WorkflowMax][]JobType{
			WorkflowPullRequest: {golangTestJobType, golangLintJobType},
			WorkflowMerge:       {golangTestJobType, golangLintJobType},
		},
	},
	{
		Identifier: "terraformtarget",
		KeyFile:    "terraform.tf",
		Workflows: [WorkflowMax][]JobType{
			WorkflowPullRequest: {
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
			WorkflowMerge: {
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

var golangLintJobType = JobType{
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

var golangTestJobType = JobType{
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
