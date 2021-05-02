package projects

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Render renders workflows into workflow YAML files in the provided output
// directory.
func Render(outDir string, workflows Workflows) error {
	for workflow, jobs := range workflows {
		workflow := WorkflowIdentifier(workflow)
		if err := RenderWorkflow(outDir, workflow, jobs); err != nil {
			return fmt.Errorf("rendering workflow %s: %w", workflow, err)
		}
	}

	return nil
}

// RenderWorkflow renders a single workflow into a workflow YAML file in the
// provided output directory.
func RenderWorkflow(outDir string, workflow WorkflowIdentifier, jobs []*Job) error {
	filePath := filepath.Join(outDir, workflow.FileName())
	jobMap := make(map[string]*Job, len(jobs))
	for _, job := range jobs {
		jobMap[job.Identifier] = job
	}
	return withFileCreate(
		filePath,
		func(file *os.File) error {
			enc := yaml.NewEncoder(file)
			enc.SetIndent(2)
			return enc.Encode(struct {
				Name string `yaml:"name"`
				On   map[string]struct {
					Branches []string `yaml:"branches"`
				} `yaml:"on"`
				Jobs map[string]*Job `yaml:"jobs"`
			}{
				Name: workflow.String(),
				On: map[string]struct {
					Branches []string `yaml:"branches"`
				}{
					workflow.Trigger(): {Branches: []string{"master"}},
				},
				Jobs: jobMap,
			})
		},
	)
}

func withFileCreate(filePath string, f func(f *os.File) error) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	if err := f(file); err != nil {
		return err
	}
	return file.Close()
}
