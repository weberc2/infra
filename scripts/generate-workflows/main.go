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

// NOTES:
// * Our golang-lambda project type needs to be able to say that it needs a
//   golang-publish project type. Specifically it needs to be able to build
//   a 'deploy' job that depends on the 'golang-publish' job's 'publish' job.
// * What part of the golang-lambda project type knows that the 'deploy' job
//   depends on the golang-publish.publish job?
// * Maybe a JobType says it needs dependencies of a particular
//   "golang-publish.publish" job type and the ProjectType contains the
//   information which routes its `golang-publish` dependency's `publish`
//   JobType

// Workflow represents the workflows in the repo.
type WorkflowIdentifier int

const (
	// WorkflowPullRequest identifies the pull request workflow.
	WorkflowIdentifierPullRequest WorkflowIdentifier = iota

	// WorkflowMerge identifies the merge workflow.
	WorkflowIdentifierMerge
)

// String returns the string-representation of a `WorkflowIdentifier`.
func (wid WorkflowIdentifier) String() string {
	switch wid {
	case WorkflowIdentifierPullRequest:
		return "pull_request"
	case WorkflowIdentifierMerge:
		return "merge"
	default:
		panic(fmt.Sprintf("Invalid workflow: %d", w))
	}
}

// JobType represents a job in a workflow. Given a `Project`, it can be used to
// instantiate concrete jobs in a workflow.
type JobType struct {
	// Identifier is the name of the job. It typically should contain
	// `${project}`, which will be interpolated with the name of the project
	// at rendering time.
	Identifier string

	Dependencies []*JobType

	// Workflow identifies the workflow to which jobs of this type belong.
	Workflow WorkflowIdentifier

	// Template is the text template that will be used to render a project into
	// a concrete job. See `ProjectTemplateParams` for the parameters available
	// to the template.
	Template *template.Template
}

// Since projects can live in any directory in the repo, but since we take the
// basename of the project directory as the project name, it is possible that
// there are two projects with the same name (e.g., `bar/foo` and `baz/foo`).
// This would allow us to have collisions between files with the same name in
// the `~/.github/workflows` directory (e.g., both `bar/foo` and `baz/foo`
// would result in two attempts to write `~/.github/workflows/build-foo.yaml`
// or similar. In order to resolve this, we will prefix the project name with
// the project type such that we can have multiple projects with the same
// basename without conflicts in the `~/.github/workflows` output directory,
// and to prevent conflicts between projects of the same project type we will
// detect and error.

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

	Assignments []JobTypeAssignment

	Dependencies []*ProjectType

	// JobType contains a list of job types associated with the project type.
	Jobs []JobType
}

// JobTypeAssignment expresses the equivalency:
// Source.Jobs[TargetJobIndex].Dependencies[TargetJobIndex] = Source.Dependencies[DependencyIndex].Jobs[DependencyJobIndex]
// From which we can validate that this expression is type-safe (both sides
// have the same type)
type JobTypeAssignment struct {
	DependencyIndex          int // The index of the dependency in Source.Dependencies
	DependencyJobIndex       int // The index of the job on the Dependency
	TargetJobIndex           int // The index of the job on the source onto which the dependency job will be mapped
	TargetJobDependencyIndex int // the dependency slot on the target job onto which the
	DependencySlot           int // The slot on the job on the dependency onto which the
}

func (pt *ProjectType) Jobs(project *Project) []Job {
	jobs := make([]Job, len(pt.Jobs))
	for i, jobType := range pt.Jobs {
		jobs[i].Dependencies = make([]string, len(jobType.Dependencies))
	}

	for _, assignment := range pt.Assignments {
		var dependencyProjectType *ProjectType = pt.Dependencies[assignment.DependencyIndex]
		var dependencyJobType *JobType = &dependencyProjectType.Jobs[assignment.DependencyJobIndex]
		var job *Job = &jobs[assignment.TargetJobIndex]
		job.Dependencies[assignment.DependencySlot] = project.Dependencies
	}
}

type ProjectIdentifier struct {
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
func (pi *ProjectIdentifier) Name() string {
	return fmt.Sprintf("%s-%s", pi.Type.Identifier, filepath.Base(pi.Path))
}

type Job struct {
	Identifier   string
	ProjectName  string
	ProjectPath  string
	Dependencies []string
	Template     *template.Template
}

// Project represents a project in a repository. Each project has a type (see
// `ProjectType` for more information) and a path (relative to the repo root).
type Project struct {
	ProjectIdentifier

	// Dependencies holds a list of names of other projects. See
	// `Project.Name()` for details on the structure/syntax of a project name.
	Dependencies []*Project
}

func (p *Project) renderTemplates(dir string) error {
	for _, template := range p.Type.Templates {
		fileName := strings.Replace(template.Name(), "${project}", p.Name(), -1)
		filePath := filepath.Join(dir, fileName)
		dependencies := make([]string, len(p.Dependencies))
		if err := func() error {
			file, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			return template.Execute(file, ProjectTemplateParams{
				Name: p.Name(),
				Path: p.Path,
			})
		}(); err != nil {
			return fmt.Errorf(
				"Applying template '%s' to project '%s': %w",
				template.Name(),
				p.Path,
				err,
			)
		}
		log.Println(success("Staged %s", fileName))
	}
	return nil
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
				"Duplicate projects detected: '%s' and '%s': Two projects "+
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
		log.Fatal(color.RedString("FATAL Creating temp dir: %v", err))
	}
	defer os.RemoveAll(tmpDir)

	// Find the root of the repository
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(color.RedString("FATAL Getting working directory: %v", err))
	}
	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		log.Fatal(color.RedString("FATAL Finding repo root: %v", err))
	}

	dir := filepath.Join(repoRoot, ".github/workflows")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Collect the projects from the repository
	projects, err := FindProjects(projectTypes, repoRoot)
	if err != nil {
		log.Fatal(color.RedString("FATAL Collecting projects: %v", err))
	}

	// Render the templates for each project
	for _, project := range projects {
		if err := project.renderTemplates(tmpDir); err != nil {
			log.Fatal(err)
		}
	}

	// Render the static files
	for fileName, contents := range staticFiles {
		filePath := filepath.Join(tmpDir, fileName)
		func() {
			file, err := os.Create(filePath)
			if err != nil {
				log.Fatal(color.RedString(
					"FATAL Creating static file '%s': %v",
					filePath,
					err,
				))
			}
			defer file.Close()

			if _, err := file.WriteString(contents); err != nil {
				log.Fatal(color.RedString(
					"FATAL Writing to static file '%s': %v",
					filePath,
					err,
				))
			}
		}()
		log.Println(success("Staged %s", fileName))
	}

	// Atomically "commit" the changes to `~/.github/workflows`.
	if err := os.Rename(tmpDir, dir); err != nil {
		if os.IsExist(err) {
			if err := os.RemoveAll(dir); err != nil {
				log.Fatalf(color.RedString(
					"FATAL Removing dir '%s': %v",
					dir,
					err,
				))
			}
			if err := os.Rename(tmpDir, dir); err != nil {
				log.Fatal(color.RedString(
					"FATAL Renaming '%s' to '%s': %v",
					tmpDir,
					dir,
					err,
				))
			}
		} else {
			log.Fatal(color.RedString(
				"FATAL Renaming '%s' to '%s': %v",
				tmpDir,
				dir,
				err,
			))
		}
	}

	log.Println(success("Promoted staged files"))
}

func makeTemplate(name, body string) *template.Template {
	return template.Must(
		template.New(name).Parse(strings.Replace(body, "\t", "    ", -1)),
	)
}

var projectTypes = []ProjectType{
	ProjectType{
		Identifier: "golanglambdasource",
		KeyFile:    "golang-lambda-source.yaml",
		Templates:  []*template.Template{},
	},
	ProjectType{
		Identifier: "golang",
		KeyFile:    "go.mod",
		Templates: []*template.Template{
			makeTemplate(
				"${project}-test.yaml",
				`name: {{ .Name }} test
on:
  pull_request:
    branches: [ master ]
  push:
    branches: [ master ]

jobs:
  {{ .Name }}-test:
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
		},
	},
	ProjectType{
		Identifier: "terraformtarget",
		KeyFile:    "terraform.tf",
		Templates: []*template.Template{
			makeTemplate(
				"${project}-plan.yaml",
				`name: {{ .Name }} plan
on:
  pull_request:
    branches: [ master ]

jobs:
  {{ .Name }}-plan-check:
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
			makeTemplate(
				"${project}-apply.yaml",
				`name: {{ .Name }} apply
on:
  push:
    branches: [ master ]

jobs:
  apply:
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
        run: pushd scripts/generate-workflows && go run .; popd
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

func success(format string, v ...interface{}) string {
	return color.GreenString("✅ "+format, v...)
}
