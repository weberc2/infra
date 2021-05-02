package projects

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type ProjectIdentifier struct {
	Path string
	Type *ProjectType
}

// Project represents a project in a repository. Each project has a type (see
// `ProjectType` for more information) and a path (relative to the repo root).
type Project struct {
	// Type is the type of a project. It is used to render the final output
	// files associated with this project into the `~/.github/workflows`
	// directory.
	Type *ProjectType

	// Path is the path to the project directory relative to the root of the
	// repository. The "name" of the project is the basename of the path
	// prefixed by the project's type identifier (see
	// `ProjectType.Identifier`).
	Path string

	Dependencies map[string]ProjectIdentifier
}

// Name returns the name of the project by appending the basename of the
// project's `Path` with the project's type's identifier. This will be used as
// a parameter to template the output files.
func (p *Project) Name() string {
	return fmt.Sprintf("%s-%s", p.Type.Identifier, filepath.Base(p.Path))
}

// FindProjects searches the repo root to locate project directories and builds
// `Project`s from them. It will return an error if multiple projects were
// detected with the same basename and type.
func FindProjects(types []ProjectType, repoRoot string) ([]Project, error) {
	projects, err := findProjects(types, repoRoot, repoRoot)
	if err != nil {
		return nil, err
	}

	sort.Slice(projects, func(i, j int) bool {
		pi, pj := projects[i], projects[j]
		return pi.Type.Identifier < pj.Type.Identifier && pi.Name() < pj.Name()
	})

	for i := range projects[:len(projects)-1] {
		pi, pj := projects[i], projects[i+1]
		if pi.Type.Identifier == pj.Type.Identifier && pi.Name() == pj.Name() {
			return nil, fmt.Errorf(
				"duplicate projects detected: '%s' and '%s': two projects "+
					"may not share the same basename and project type",
				pi.Path,
				pj.Path,
			)
		}
	}

	return projects, nil
}

// findProjects is a recursive helper for `FindProjects`.
func findProjects(types []ProjectType, root, dir string) ([]Project, error) {
	parser := projectParser{types: types, repoRoot: root}
	err := parser.parseProjectsRecursive(dir)
	return parser.projects, err
}

type projectParser struct {
	types    []ProjectType
	projects []Project
	repoRoot string
}

func (pp *projectParser) parseProjectsRecursive(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() == keyFileName {
			log.Debugf("parsing projects directory %s", dir)
			if err := pp.parseProjectsDirectory(dir); err != nil {
				return fmt.Errorf("Parsing project(s) directory '%s': %s", dir, err)
			}

			// for now, we will prohibit nested projects
			continue
		}

		if file.IsDir() {
			filePath := filepath.Join(dir, file.Name())
			if err := pp.parseProjectsRecursive(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pp *projectParser) pushProject(p Project) {
	pp.projects = append(pp.projects, p)
}

func (pp *projectParser) parseProjectsDirectory(dir string) error {
	filePath := filepath.Join(dir, keyFileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	var payload struct {
		Projects []struct {
			Type         string `yaml:"type"`
			Dependencies map[string]struct {
				Path string `yaml:"path"`
				Type string `yaml:"type"`
			} `yaml:"dependencies"`
		} `yaml:"projects"`
	}
	if err := yaml.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("Parsing YAML file '%s': %w", filePath, err)
	}

	path, err := filepath.Rel(pp.repoRoot, dir)
	if err != nil {
		return err
	}

	for _, project := range payload.Projects {
		projectType, err := pp.findType(project.Type)
		if err != nil {
			return err
		}

		dependencies := make(map[string]ProjectIdentifier, len(project.Dependencies))
		for dependencyName, dependency := range project.Dependencies {
			if dependencyType, found := projectType.Dependencies[dependencyName]; found {
				if dependencyType.Identifier != dependency.Type {
					return fmt.Errorf(
						"expected type '%s' for dependency '%s' of "+
							"(path=%s, type=%s); found type '%s'",
						dependencyType.Identifier,
						path,
						project.Type,
						dependencyName,
						dependency.Type,
					)
				}
				dependencies[dependencyName] = ProjectIdentifier{
					Path: path,
					Type: dependencyType,
				}
				continue
			}
			return fmt.Errorf(
				"Unknown dependency '%s' for project type '%s'",
				dependencyName,
				project.Type,
			)
		}

		log.Debugf(
			"adding project (path=%s, type=%s)",
			path,
			projectType.Identifier,
		)
		pp.pushProject(Project{
			Type:         projectType,
			Path:         path,
			Dependencies: dependencies,
		})
	}
	return nil
}

func (pp *projectParser) findType(identifier string) (*ProjectType, error) {
	for i := range pp.types {
		if pp.types[i].Identifier == identifier {
			return &pp.types[i], nil
		}
	}
	return nil, fmt.Errorf("project type '%s' not found", identifier)
}

// RenderProjectWorkflows collects projects in the repository, builds workflows,
// and writes workflow YAML files to disk at `outDir`.
func RenderProjectWorkflows(
	projectTypes []ProjectType,
	repoRoot string,
	outDir string,
) error {
	projects, err := FindProjects(projectTypes, repoRoot)
	if err != nil {
		return fmt.Errorf("Collecting projects: %w", err)
	}

	workflows, err := MaterializeWorkflows(projects)
	if err != nil {
		return fmt.Errorf("Building workflows: %w", err)
	}

	if err := Render(outDir, workflows); err != nil {
		return fmt.Errorf("Rendering workflows: %w", err)
	}

	return nil
}

const keyFileName = "projects.yaml"
