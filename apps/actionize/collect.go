package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const projectsFileName = "projects.yaml"

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
		Projects []Project `yaml:"projects"`
	}
	if err := yaml.Unmarshal(data, &projects); err != nil {
		return err
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
