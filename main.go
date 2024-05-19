package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"text/template"
)

var reOverlays = regexp.MustCompile(`^.*overlays$`)
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
	appDir := flag.String("app", "applications", "The directory that contains the applications")
	env := flag.String("env", "beta", "The environment stage")
	//	exclude := flag.String("exclude", "", "The service names to exclude")
	project := flag.String("project", "", "The name of the project")
	root := flag.String("root", ".", "The root of the project directory")
	flag.Parse()
	if *project == "" {
		ifErr(errors.New("No `project` given, exiting."))
	}
	services := []string{}
	envConfigPath := fmt.Sprintf("%s/config/%s", *root, *env)
	_, err := os.Stat(envConfigPath)
	if errors.Is(err, os.ErrNotExist) {
		ifErr(os.MkdirAll(envConfigPath, os.ModePerm))
	}
	filePath := fmt.Sprintf("%s/kustomization.yaml", envConfigPath)
	_, err = os.Stat(filePath)
	if !errors.Is(err, os.ErrNotExist) {
		services, err = getCurrentEntries(filePath)
		ifErr(err)
	}
	fileSystem := os.DirFS(fmt.Sprintf("%s/%s/%s", *root, *appDir, *project))
	fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		ifErr(err)
		if reOverlays.MatchString(path) {
			envDir := fmt.Sprintf("%s/%s/%s", fileSystem, path, *env)
			_, err := os.Stat(envDir)
			if !errors.Is(err, os.ErrNotExist) {
				// The entries MUST be created relative to the repository root for `kustomize` to work.
				relativeEnvDir := fmt.Sprintf("../.%s", envDir)
				if !slices.Contains(services, relativeEnvDir) {
					services = append(services, relativeEnvDir)
				}
			}
		}
		return nil
	})
	if len(services) > 0 {
		fd, err := os.Create(filePath)
		ifErr(err)
		tpl := template.Must(template.New("kustomization.yaml").Parse(tplKustomization))
		err = tpl.Execute(fd, Project{
			Name:      *project,
			Env:       *env,
			Resources: services,
		})
		ifErr(err)
	}
}
