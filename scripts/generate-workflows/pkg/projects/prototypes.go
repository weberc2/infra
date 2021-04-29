package projects

import (
	"fmt"
	"text/template"
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
