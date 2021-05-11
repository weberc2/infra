package main

import (
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

func (pji *ProjectJobIdentifier) UnmarshalYAML(value *yaml.Node) error {
	const errFormat = "job identifiers must be strings with form " +
		"'{project}:{job}'; %s"
	switch value.Kind {
	case yaml.ScalarNode:
		chunks := strings.SplitN(value.Value, ":", 1)
		if len(chunks) < 2 {
			return fmt.Errorf(errFormat, "missing delimiter ':'")
		} else if len(chunks[0]) < 1 {
			return fmt.Errorf(
				errFormat,
				"{project} must be at least 1 character",
			)
		} else if len(chunks[1]) < 1 {
			return fmt.Errorf(
				errFormat,
				"{job} must be at least 1 character",
			)
		}
		pji.Project = ProjectName(chunks[0])
		pji.Job = ProjectJobName(chunks[1])
		return nil
	default:
		return fmt.Errorf(
			errFormat,
			fmt.Sprintf("'%v' is not a string", value.Kind),
		)
	}
}

type ProjectJob struct {
	Name   ProjectJobName         `yaml:"name"`
	RunsOn string                 `yaml:"runs-on"`
	Needs  []ProjectJobIdentifier `yaml:"needs,omitempty"`
	Steps  []Step                 `yaml:"steps"`
}

type Project struct {
	Name ProjectName                   `yaml:"name"`
	Path string                        `yaml:"path"`
	Jobs map[WorkflowName][]ProjectJob `yaml:"jobs"`
}
