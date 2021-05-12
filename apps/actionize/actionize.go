package main

import "fmt"

func buildWorkflows(projects []Project) []Workflow {
	pv := projectVisitor{workflows: map[WorkflowName][]Job{}}
	for i := range projects {
		pv.visitProject(&projects[i])
	}
	return pv.actionize()
}

type projectVisitor struct {
	workflows map[WorkflowName][]Job
}

func (pv *projectVisitor) visitProject(p *Project) {
	for workflowName, projectJobs := range p.Jobs {
		jobs := make([]Job, len(projectJobs))
		for i := range projectJobs {
			jobs[i] = projectJobs[i].Actionize(p.Name)
		}
		pv.workflows[workflowName] = append(pv.workflows[workflowName], jobs...)
	}
}

func (pv *projectVisitor) actionize() []Workflow {
	workflows := make([]Workflow, 0, len(pv.workflows))
	for workflowName, jobs := range pv.workflows {
		workflows = append(workflows, Workflow{Name: workflowName, Jobs: jobs})
	}
	return workflows
}

func (pji *ProjectJobIdentifier) Actionize() JobName {
	return JobName(fmt.Sprintf("%s_%s", pji.Project, pji.Job))
}

func (pj *ProjectJob) Actionize(projectName ProjectName) Job {
	needs := make([]JobName, len(pj.Needs))
	for i := range pj.Needs {
		needs[i] = pj.Needs[i].Actionize()
	}
	return Job{
		Name:   JobName(fmt.Sprintf("%s_%s", projectName, pj.Name)),
		RunsOn: pj.RunsOn,
		Needs:  needs,
		Steps:  pj.Steps,
	}
}
