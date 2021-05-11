package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func renderWorkflows(outputDirectory string, workflows []Workflow) error {
	for i := range workflows {
		filePath := filepath.Join(
			outputDirectory,
			workflows[i].Name.FileName(),
		)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		enc := yaml.NewEncoder(file)
		enc.SetIndent(2)
		err = enc.Encode(&workflows[i])
		if err := file.Close(); err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func renderWorkflowsTx(outputDirectory string, workflows []Workflow) error {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if err := renderWorkflows(tempDir, workflows); err != nil {
		return err
	}
	if err := os.RemoveAll(outputDirectory); err != nil {
		return err
	}
	return os.Rename(tempDir, outputDirectory)
}
