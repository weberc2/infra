package main

import (
	"fmt"
	"text/template"
)

type Job struct {
	Identifier  string
	Name        string
	ProjectName string
	ProjectPath string
	Template    *template.Template
}

type Workflows [WorkflowMax][]Job

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
