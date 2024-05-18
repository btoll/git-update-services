package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"text/template"
)

var reResources = regexp.MustCompile(`-\s(?P<Resource>.*$)`)

var tplKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  {{ range .Resources -}}
  - {{ . }}
  {{ end -}}`

type Project struct {
	Name      string
	Env       string
	Resources []string
}

func getCurrentEntries(filename string) ([]string, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	current := []string{}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if reResources.MatchString(line) {
			matches := reResources.FindStringSubmatch(line)
			current = append(current, matches[1])
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return current, nil
}

func ifErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	env := flag.String("env", "beta", "The environment stage")
	//	exclude := flag.String("exclude", "", "The service names to exclude")
	project := flag.String("project", "", "The name of the project")
	root := flag.String("root", ".", "The root of the project directory")
	flag.Parse()

	services := []string{}
	filePath := fmt.Sprintf("%s/config/%s/kustomization.yaml", *root, *env)
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("No `kustomization.yaml` file, creating.")
	} else {
		services, err = getCurrentEntries(filePath)
		ifErr(err)
	}
	entries, err := os.ReadDir(fmt.Sprintf("%s/applications/%s", *root, *project))
	ifErr(err)
	for _, entry := range entries {
		if entry.IsDir() {
			// The entries MUST be created relative to the repository root for `kustomize` to work.
			newResource := fmt.Sprintf("../../applications/%s/%s/overlays/%s", *project, entry.Name(), *env)
			if !slices.Contains(services, newResource) {
				services = append(services, newResource)
			}
		}
	}
	p := Project{
		Name:      *project,
		Env:       *env,
		Resources: services,
	}
	fd, err := os.Create(filePath)
	ifErr(err)
	tpl := template.Must(template.New("kustomization.yaml").Parse(tplKustomization))
	err = tpl.Execute(fd, p)
	ifErr(err)
}
