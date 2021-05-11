package main

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type WorkflowName int

func (wfn *WorkflowName) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("wanted workflow name (string) found yaml node with kind = %v", value.Kind)
	}
	return wfn.FromString(value.Value)
}

func (wfn *WorkflowName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("wanted workflow name (string): %w", err)
	}
	return wfn.FromString(s)
}

const (
	WorkflowPullRequest WorkflowName = iota
	WorkflowMerge
)

func (wfn *WorkflowName) FromString(s string) error {
	switch s {
	case "pull-request":
		*wfn = WorkflowPullRequest
		return nil
	case "merge":
		*wfn = WorkflowMerge
		return nil
	default:
		return fmt.Errorf(
			"invalid workflow name '%s'; wanted 'pull-request' or 'merge'",
			s,
		)
	}
}

func (wfn WorkflowName) String() string {
	switch wfn {
	case WorkflowPullRequest:
		return "Pull Request"
	case WorkflowMerge:
		return "Merge"
	default:
		panic(fmt.Sprintf("Invalid workflow name: %d", wfn))
	}
}

func (wfn WorkflowName) FileName() string {
	switch wfn {
	case WorkflowPullRequest:
		return "pull-request.yaml"
	case WorkflowMerge:
		return "merge.yaml"
	default:
		panic(fmt.Sprintf("Invalid workflow name: %d", wfn))
	}
}

func (wfn WorkflowName) TriggerString() string {
	switch wfn {
	case WorkflowPullRequest:
		return "pull_request"
	case WorkflowMerge:
		return "push"
	default:
		panic(fmt.Sprintf("Invalid workflow name: %d", wfn))
	}
}

type Step struct {
	Name string            `json:"name,omitempty" yaml:"name,omitempty"`
	Env  map[string]string `json:"env,omitempty"  yaml:"env,omitempty"`
	Uses string            `json:"uses,omitempty" yaml:"uses,omitempty"`
	Run  string            `json:"run,omitempty"  yaml:"run,omitempty"`
}

type JobName string

type Job struct {
	Name   JobName
	RunsOn string
	Needs  []JobName
	Steps  []Step
}

func (j *Job) YAMLNode() *yaml.Node {
	nodes := []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "runs-on"},
		{Kind: yaml.ScalarNode, Value: j.RunsOn},
	}
	if len(j.Needs) > 0 {
		needsNodes := make([]*yaml.Node, len(j.Needs))
		for i, need := range j.Needs {
			needsNodes[i] = &yaml.Node{Kind: yaml.ScalarNode, Value: string(need)}
		}
		nodes = append(
			nodes,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "needs"},
			&yaml.Node{Kind: yaml.SequenceNode, Content: needsNodes},
		)
	}

	if len(j.Steps) > 0 {
		stepNodeValues := make([]yaml.Node, len(j.Steps))
		stepNodes := make([]*yaml.Node, len(j.Steps))
		for i, step := range j.Steps {
			if err := stepNodeValues[i].Encode(&step); err != nil {
				panic(fmt.Sprintf("Error marshaling Step to YAML: %v", err))
			}
			if err := validate(&stepNodeValues[i]); err != nil {
				panic(err)
			}
			stepNodes[i] = &stepNodeValues[i]
		}
		nodes = append(
			nodes,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "steps"},
			&yaml.Node{Kind: yaml.SequenceNode, Content: stepNodes},
		)
	}

	return &yaml.Node{Kind: yaml.MappingNode, Content: nodes}
}

type Workflow struct {
	Name WorkflowName
	Jobs []Job
}

func (wf *Workflow) MarshalYAML() (interface{}, error) {
	jobNodes := make([]*yaml.Node, 2*len(wf.Jobs))
	for i, job := range wf.Jobs {
		jobNodes[2*i] = &yaml.Node{Kind: yaml.ScalarNode, Value: string(job.Name)}
		jobNodes[2*i+1] = job.YAMLNode()
	}

	out := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: wf.Name.String()},
			{Kind: yaml.ScalarNode, Value: "on"},
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: wf.Name.TriggerString()},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "branches"},
							{
								Kind:  yaml.SequenceNode,
								Style: yaml.FlowStyle,
								Content: []*yaml.Node{{
									Kind:  yaml.ScalarNode,
									Value: "40-lambda-support",
								}},
							},
						},
					},
				},
			},
			{Kind: yaml.ScalarNode, Value: "jobs"},
			{Kind: yaml.MappingNode, Content: jobNodes},
		},
	}
	return out, nil
}

func validate(n *yaml.Node) error {
	if n == nil {
		return fmt.Errorf("node is nil")
	}
	var name string
	switch n.Kind {
	case 0:
		name = "zero"
	case yaml.AliasNode:
		name = "alias"
	case yaml.DocumentNode:
		name = "document"
	case yaml.ScalarNode:
		name = "scalar"
	case yaml.MappingNode:
		name = "mapping"
	case yaml.SequenceNode:
		name = "sequence"
	default:
		panic("Unknown kind")
	}
	for i, contentNode := range n.Content {
		if err := validate(contentNode); err != nil {
			return fmt.Errorf("validating %v node's child (index %d): %w", name, i, err)
		}
	}
	return nil
}
