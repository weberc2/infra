package projects

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
func RenderWorkflow(outDir string, workflow WorkflowIdentifier, jobs []Job) error {
	filePath := filepath.Join(outDir, workflow.FileName())
	return withFileCreate(
		filePath,
		func(file *os.File) error {
			if _, err := fmt.Fprintf(file, "name: %s\n", workflow); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(
				file,
				"on:\n  %s:\n    branches: [ master ]\n\njobs:\n",
				workflow.Trigger(),
			); err != nil {
				return err
			}

			for i := range jobs {
				if _, err := file.WriteString("\n"); err != nil {
					return err
				}
				if err := renderJob(file, &jobs[i]); err != nil {
					return fmt.Errorf(
						"rendering job '%s' in workflow '%s': %w",
						jobs[i].Name,
						workflow,
						err,
					)
				}
			}

			return nil
		},
	)
}

func renderJob(file *os.File, job *Job) error {
	var writer strings.Builder

	if err := job.Template.Execute(&writer, struct {
		Name string
		Path string
	}{
		Name: job.ProjectName,
		Path: job.ProjectPath,
	}); err != nil {
		return err
	}

	// Indent all lines
	contents := strings.ReplaceAll(writer.String(), "\n", "\n  ")

	// If we indented any previously empty lines, unindent them.
	contents = strings.ReplaceAll(contents, "\n  \n", "\n\n")

	_, err := file.WriteString("  " + contents)
	return err
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
