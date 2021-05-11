package main

import (
	"fmt"
	"html/template"
	"strings"
)

func (p *Project) TemplateMut() error {
	for workflowName, jobs := range p.Jobs {
		for i := range jobs {
			for j := range jobs[i].Steps {
				if err := templateStepMut(&jobs[i].Steps[j], struct {
					Name ProjectName
					Path string
				}{
					p.Name,
					p.Path,
				}); err != nil {
					return fmt.Errorf(
						"templating job '%s' in workflow '%s' in project "+
							"'%s' at path '%s': %w",
						jobs[i].Name,
						workflowName,
						p.Name,
						p.Path,
						err,
					)
				}
			}
		}
	}
	return nil
}

func templateStepMut(step *Step, v interface{}) error {
	t, err := template.New("").Parse(step.Run)
	if err != nil {
		return err
	}
	var sb strings.Builder
	if err := t.Execute(&sb, v); err != nil {
		return err
	}
	step.Run = sb.String()
	return nil
}
