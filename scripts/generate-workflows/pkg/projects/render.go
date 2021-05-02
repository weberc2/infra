package projects

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Render renders workflows into workflow YAML files in the provided output
// directory.
func Render(outDir string, workflows []Workflow) error {
	for i := range workflows {
		if err := RenderWorkflow(outDir, &workflows[i]); err != nil {
			return fmt.Errorf(
				"rendering workflow %s: %w",
				workflows[i].Identifier,
				err,
			)
		}
	}

	return nil
}

// RenderWorkflow renders a single workflow into a workflow YAML file in the
// provided output directory.
func RenderWorkflow(outDir string, workflow *Workflow) error {
	return withFileCreate(
		filepath.Join(outDir, workflow.Identifier.FileName()),
		func(file *os.File) error {
			enc := yaml.NewEncoder(file)
			enc.SetIndent(2)
			return enc.Encode(workflow)
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
