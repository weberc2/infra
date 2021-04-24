package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	dir := "./.github/workflows"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("Creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targets, err := findTargets("./targets")
	if err != nil {
		log.Fatalf("Finding targets: %v", err)
	}

	for i := range targets {
		t, err := filepath.Rel("./targets", targets[i])
		if err != nil {
			log.Fatalf("Finding relative path of '%s': %v", targets[i], err)
		}
		targets[i] = t
	}

	for _, target := range targets {
		for templateName, template := range templates {
			if err := applyTemplate(
				tmpDir,
				target,
				templateName,
				template,
			); err != nil {
				log.Fatalf("Applying template %s: %v", templateName, err)
			}
		}
	}

	for fileName, contents := range staticFiles {
		filePath := filepath.Join(dir, fileName)
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("Creating static file '%s': %v", filePath, err)
		}
		defer file.Close()

		if _, err := file.WriteString(contents); err != nil {
			log.Fatalf("Writing to static file '%s': %v", filePath, err)
		}
	}

	if err := os.Rename(tmpDir, dir); err != nil {
		if os.IsExist(err) {
			if err := os.RemoveAll(dir); err != nil {
				log.Fatalf("Removing dir '%s': %v", dir, err)
			}
			if err := os.Rename(tmpDir, dir); err != nil {
				log.Fatalf("Renaming '%s' to '%s': %v", tmpDir, dir, err)
			}
		} else {
			log.Fatalf("Renaming '%s' to '%s': %v", tmpDir, dir, err)
		}
	}
}

func findTargets(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var targets []string
	for _, file := range files {
		if file.IsDir() {
			nestedTargets, err := findTargets(filepath.Join(dir, file.Name()))
			if err != nil {
				return nil, err
			}
			targets = append(targets, nestedTargets...)
		} else {
			if file.Name() == "terraform.tf" {
				return []string{dir}, nil
			}
		}
	}

	return targets, nil
}

func applyTemplate(dir string, target string, templateName string, template *template.Template) error {
	filePath := filepath.Join(
		dir,
		strings.Replace(templateName, "${target}", target, -1),
	)
	if err := os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return template.Execute(file, struct{ Target string }{target})
}

var templates = map[string]*template.Template{
	"${target}-plan.yml": template.Must(
		template.New("plan").Parse(
			`name: {{ .Target }} plan
on:
  pull_request:
    branches: [ master ]

jobs:
  {{ .Target }}-plan-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Terraform setup
        uses: hashicorp/setup-terraform@v1
      - name: Terraform init
        env:
          AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
        run: terraform -chdir=./targets/{{ .Target }} init
      - name: Terraform plan
        env:
          AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
        run: terraform -chdir=./targets/{{ .Target }} plan
`,
		),
	),
	"${target}-apply.yml": template.Must(
		template.New("plan").Parse(
			`name: {{ .Target }} Test
on:
  push:
    branches: [ master ]

jobs:
  apply:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v2
      - name: Terraform setup
        uses: hashicorp/setup-terraform@v1
      - name: Terraform init {{ .Target }}
        id: init-{{ .Target }}
        env:
          AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
        run: terraform -chdir=./targets/{{ .Target }} init
      - name: Terraform apply {{ .Target }}
        id: apply-{{ .Target }}
        env:
          AWS_ACCESS_KEY_ID: ${{"{{"}} secrets.TERRAFORM_AWS_ACCESS_KEY_ID {{"}}"}}
          AWS_SECRET_ACCESS_KEY: ${{"{{"}} secrets.TERRAFORM_AWS_SECRET_ACCESS_KEY {{"}}"}}
        run: terraform -chdir=./targets/{{ .Target }} apply -auto-approve
`,
		),
	),
}

var staticFiles = map[string]string{
	"terraform-fmt.yml": `name: Terraform format check

on:
  pull_request:
    branches: [ master ]

jobs:
  format-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Terraform setup
        uses: hashicorp/setup-terraform@v1
      - name: Terraform format check
        run: terraform fmt -recursive -check
`,
}
