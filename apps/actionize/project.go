package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProjectName string

type ProjectJobName string

type ProjectJobIdentifier struct {
	Project ProjectName
	Job     ProjectJobName
}

func (pji *ProjectJobIdentifier) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Unmarshaling ProjectJobIdentifier: %w", err)
	}
	return pji.FromString(s)
}

func (pji *ProjectJobIdentifier) UnmarshalYAML(value *yaml.Node) error {
	const errFormat = "job identifiers must be strings with form " +
		"'{project}:{job}'; %s"
	switch value.Kind {
	case yaml.ScalarNode:
		return pji.FromString(value.Value)
	default:
		return fmt.Errorf(
			errFormat,
			fmt.Sprintf("'%v' is not a string", value.Kind),
		)
	}
}

func (pji *ProjectJobIdentifier) FromString(s string) error {
	const errFormat = "job identifiers must be strings with form " +
		"'{project}:{job}' (got '%s'); %s"
	chunks := strings.SplitN(s, ":", 2)
	if len(chunks) != 2 {
		return fmt.Errorf(errFormat, s, "missing delimiter ':'")
	} else if len(chunks[0]) < 1 {
		return fmt.Errorf(
			errFormat,
			s,
			"{project} must be at least 1 character",
		)
	} else if len(chunks[1]) < 1 {
		return fmt.Errorf(
			errFormat,
			s,
			"{job} must be at least 1 character",
		)
	}
	pji.Project = ProjectName(chunks[0])
	pji.Job = ProjectJobName(chunks[1])
	return nil
}

type ProjectJob struct {
	Name   ProjectJobName         `json:"name" yaml:"name"`
	RunsOn string                 `json:"runs-on" yaml:"runs-on"`
	Needs  []ProjectJobIdentifier `json:"needs,omitempty" yaml:"needs,omitempty"`
	Steps  []Step                 `json:"steps" yaml:"steps"`
}

type Project struct {
	Name ProjectName                   `json:"name" yaml:"name"`
	Path string                        `json:"path" yaml:"path"`
	Jobs map[WorkflowName][]ProjectJob `json:"jobs" yaml:"jobs"`
}

func (p *Project) UnmarshalJSON(data []byte) error {
	var p2 struct {
		Name ProjectName             `json:"name"`
		Path string                  `json:"path"`
		Jobs map[string][]ProjectJob `json:"jobs"`
	}
	if err := json.Unmarshal(data, &p2); err != nil {
		return err
	}
	jobs := make(map[WorkflowName][]ProjectJob, len(p2.Jobs))
	for key, value := range p2.Jobs {
		var wfn WorkflowName
		if err := wfn.FromString(key); err != nil {
			return err
		}
		jobs[wfn] = value
	}
	p.Name = p2.Name
	p.Path = p2.Path
	p.Jobs = jobs
	return nil
}
