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

	Dependencies []string

	// Template is the text template used to generate the job's YAML.  Note that
	// the template's source text *should not* be indented, but rather the
	// output text will be indented automatically before being attached to the
	// output workflow file.
	Template *template.Template
}

// Workflows is a data structure that associates a list of jobs to a workflow.
// It's an array of length `WorkflowMax` and it's intended to keyed by valid
// `WorkflowIdentifier`s.
type Workflows [WorkflowMax][]*Job

// MaterializeWorkflows takes a list of projects and returns the corresponding
// workflows.
func MaterializeWorkflows(projects []Project) (Workflows, error) {
	m := materializer{cache: map[cacheKey]int{}, projects: projects}
	return m.materializeWorkflows()
}

type cacheKey struct {
	workflow              WorkflowIdentifier
	projectTypeIdentifier string
	projectPath           string
	jobTypeName           string
}

type materializer struct {
	cache     map[cacheKey]int
	workflows Workflows
	projects  []Project
}

func (m *materializer) materializeWorkflows() (Workflows, error) {
	for _, project := range m.projects {
		for workflowIdentifier, jobTypes := range project.Type.Workflows {
			for i := range jobTypes {
				if _, err := m.materializeJob(
					WorkflowIdentifier(workflowIdentifier),
					&jobTypes[i],
					&project,
				); err != nil {
					return Workflows{}, err
				}
			}
		}
	}

	return m.workflows, nil
}

func (m *materializer) materializeJob(
	workflow WorkflowIdentifier,
	jobType *JobType,
	parentProject *Project,
) (*Job, error) {
	key := cacheKey{
		workflow:              workflow,
		projectTypeIdentifier: parentProject.Type.Identifier,
		projectPath:           parentProject.Path,
		jobTypeName:           jobType.Name,
	}

	idx, found := m.cache[key]
	if found {
		return m.workflows[workflow][idx], nil
	}

	m.cache[key] = len(m.workflows[workflow])
	dependencies := make([]string, len(jobType.Dependencies))
	for i, jobDependency := range jobType.Dependencies {
		pid, found := parentProject.Dependencies[jobDependency.Name]
		if !found {
			return nil, fmt.Errorf(
				"projects of type '%s' must have dependency called '%s', but "+
					"no such dependency exists on project '%s'",
				parentProject.Type.Identifier,
				jobDependency.Name,
				parentProject.Name(),
			)
		}
		p, err := m.findProject(pid)
		if err != nil {
			return nil, fmt.Errorf(
				"looking for dependency of project (path=%s, type=%s): %w",
				pid.Path,
				pid.Type.Identifier,
				err,
			)
		}
		d, err := m.materializeJob(
			workflow,
			&parentProject.Type.Dependencies[jobDependency.Name].Workflows[workflow][jobDependency.JobIndex],
			p,
		)
		if err != nil {
			return nil, err
		}
		dependencies[i] = d.Identifier
	}

	m.workflows[workflow] = append(
		m.workflows[workflow],
		&Job{
			Identifier:   fmt.Sprintf("%s-%s", parentProject.Name(), jobType.Name),
			Name:         fmt.Sprintf("%s %s", parentProject.Name(), jobType.Name),
			ProjectName:  parentProject.Name(),
			ProjectPath:  parentProject.Path,
			Dependencies: dependencies,
			Template:     jobType.Template,
		},
	)
	return m.workflows[workflow][len(m.workflows[workflow])-1], nil
}

func (m *materializer) findProject(id ProjectIdentifier) (*Project, error) {
	for i := range m.projects {
		if m.projects[i].Path == id.Path && m.projects[i].Type.Identifier == id.Type.Identifier {
			return &m.projects[i], nil
		}
	}

	return nil, fmt.Errorf(
		"project not found (path=%s, type=%s)",
		id.Path,
		id.Type.Identifier,
	)
}
