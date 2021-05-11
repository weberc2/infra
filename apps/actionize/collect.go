package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/load"
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
	filePath := filepath.Join(dir, projectsFileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	var projects struct {
		Projects []Project `json:"projects" yaml:"projects"`
	}

	instances := cue.Build(load.Instances(
		[]string{filePath},
		&load.Config{
			ModuleRoot: pc.root,
			Package:    filepath.Base(dir),
			Dir:        dir,
		},
	))
	for _, instance := range instances {
		if instance.Err != nil {
			return fmt.Errorf("Building instance %s: %w", instance.DisplayName, instance.Err)
		}
	}

	if len(instances) != 1 {
		panic(fmt.Sprintf("Expected 1 Cue instance; found %d", len(instances)))
	}

	instance := instances[0]
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
