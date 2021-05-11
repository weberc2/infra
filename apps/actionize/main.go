package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	if err := entrypoint(); err != nil {
		log.Fatal(err)
	}
}

func entrypoint() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	root, err := findRoot(".git", wd)
	if err != nil {
		return fmt.Errorf("finding project root: %w", err)
	}

	projects, err := collectProjects(root)
	if err != nil {
		return fmt.Errorf("collecting projects: %w", err)
	}

	for i := range projects {
		if err := projects[i].TemplateMut(); err != nil {
			return err
		}
	}

	workflows := buildWorkflows(projects)
	return renderWorkflowsTx(
		filepath.Join(root, ".github/workflows"),
		workflows,
	)
}

func findRoot(fileName string, dir string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, fileName)); err == nil {
		return dir, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// If we get here, then we didn't find the file called `fileName` in the
	// directory, so we should recurse into the parent directory (assuming we're
	// not at root).
	if dir == "/" || dir == "" || filepath.Dir(dir) == dir {
		return "", fmt.Errorf(
			"Failed to find '%s'; the current directory does not appear to "+
				"be a workspace",
			fileName,
		)
	}
	return findRoot(fileName, filepath.Dir(dir))
}
