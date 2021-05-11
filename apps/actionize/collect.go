package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"cuelang.org/go/cue"
)

const projectsFileName = "projects.cue"

func collectProjects(root string) ([]Project, error) {
	pc := projectCollector{root: root}
	err := pc.collectRecursive(root)
	return pc.projects, err
}

type projectCollector struct {
	root     string
	projects []Project
}

func (pc *projectCollector) collectRecursive(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			if err := pc.collectRecursive(filepath.Join(
				dir,
				file.Name(),
			)); err != nil {
				return err
			}
			continue
		}

		if file.Name() == projectsFileName {
			if err := pc.collect(dir); err != nil {
				return fmt.Errorf(
					"unmarshaling projects from %s/%s: %w",
					dir,
					projectsFileName,
					err,
				)
			}
		}
	}

	return nil
}

func (pc *projectCollector) collect(dir string) error {
	data, err := ioutil.ReadFile(filepath.Join(dir, projectsFileName))
	if err != nil {
		return err
	}
	var projects struct {
		Projects []Project `json:"projects" yaml:"projects"`
	}

	var r cue.Runtime
	instance, err := r.Compile(projectsFileName, data)
	if err != nil {
		return fmt.Errorf("compiling '%s': %w", filepath.Join(dir, projectsFileName), err)
	}
	data, err = json.MarshalIndent(instance.Value(), "", "    ")
	if err != nil {
		return fmt.Errorf("Marhsalling '%s' to JSON:", projectsFileName, err)
	}
	if err := json.Unmarshal(data, &projects); err != nil {

		return fmt.Errorf("decoding '%s': %w", projectsFileName, err)
	}

	for i := range projects.Projects {
		path, err := filepath.Rel(pc.root, dir)
		if err != nil {
			return err
		}
		projects.Projects[i].Path = path
	}
	pc.projects = append(pc.projects, projects.Projects...)
	return nil
}
