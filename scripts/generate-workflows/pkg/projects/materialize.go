package projects

import (
	"fmt"
	"text/template"
)

// Job represents a concrete GitHub Actions job.  It has everything it needs to
// be rendered onto a GitHub Actions workflow YAML file.
type Job struct {
	// Identifier identifies the job within the workflow.
	Identifier string

	// Name is the human-readable name for the job.
	Name string

	// ProjectName is the name of the project associated with the job.
	ProjectName string

	// ProjectPath is the repo-relative path to the project associated with the
	// job.
	ProjectPath string

	// Template is the text template used to generate the job's YAML.  Note that
	// the template's source text *should not* be indented, but rather the
	// output text will be indented automatically before being attached to the
	// output workflow file.
	Template *template.Template
}

// Workflows is a data structure that associates a list of jobs to a workflow.
// It's an array of length `WorkflowMax` and it's intended to keyed by valid
// `WorkflowIdentifier`s.
type Workflows [WorkflowMax][]Job

// MaterializeWorkflows takes a list of projects and returns the corresponding
// workflows.
func MaterializeWorkflows(projects []Project) Workflows {
	var workflows Workflows

	for _, project := range projects {
		for workflowIdentifier, jobTypes := range project.Type.Workflows {
			jobs := make([]Job, len(jobTypes))
			for i, jobType := range jobTypes {
				jobs[i] = Job{
					Identifier:  fmt.Sprintf("%s-%s", project.Name(), jobType.Name),
					Name:        fmt.Sprintf("%s %s", project.Name(), jobType.Name),
					ProjectName: project.Name(),
					ProjectPath: project.Path,
					Template:    jobType.Template,
				}
			}
			workflows[workflowIdentifier] = append(workflows[workflowIdentifier], jobs...)
		}
	}

	return workflows
}
