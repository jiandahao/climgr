package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	// "gopkg.in/yaml.v3"
	yaml "github.com/goccy/go-yaml"
)

var goback2Prev = &Resource{Name: "⏎ parent ⏎", Description: "go back to previous menu"}

var configPath string = getConfigPath()

func getConfigPath() string {
	configPath := os.Getenv("CLIMGR_CONFIG_PATH")
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		configPath = filepath.Join(homeDir, ".climgr/config.d")
	}

	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		panic(fmt.Errorf("failed to create folder: %s, %v", configPath, err))
	}

	return configPath
}

func main() {
	var resources []*Resource

	entries, err := os.ReadDir(configPath)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		if !(filepath.Ext(entry.Name()) == ".yml" ||
			filepath.Ext(entry.Name()) == ".yaml") {
			continue
		}

		fd, err := os.Open(filepath.Join(configPath, entry.Name()))
		if err != nil {
			fmt.Println(err)
			return
		}

		cfgBody, err := ioutil.ReadAll(fd)
		if err != nil {
			fmt.Println(err)
			return
		}

		var resource []*Resource

		if err := yaml.Unmarshal(cfgBody, &resource); err != nil {
			fmt.Printf("failed to parse resource in %s, %v", filepath.Join(configPath, entry.Name()), err)
			return
		}

		resources = append(resources, resource...)
	}

	menu := &ResourceMenu{
		Items: resources,
	}

	if err := menu.ShowPanel(); err != nil {
		fmt.Println(err)
	}
}

var registeredActions = map[string]func(res *Resource) error{
	"":      func(res *Resource) error { return nil }, // do nothing
	"print": func(res *Resource) error { fmt.Println(res.Description); return nil },
	"run": func(res *Resource) error {
		if res.Command == "" {
			fmt.Println("Not command found")
			return nil
		}

		fmt.Println(res.Command)
		cmd := exec.Command("bash", "-c", res.Command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	},
}

// Resource describes a selectable resource
type Resource struct {
	Name           string      `yaml:"name"`           // 资源名称
	Command        string      `yaml:"command"`        // 命令，当SelectedAction为run时，将自动执行
	Description    string      `yaml:"description"`    // 描述
	SelectedAction string      `yaml:"selectedAction"` // 选中后的动作
	Children       []*Resource `yaml:"children"`       // 子资源
}

// ResourceMenu 资源菜单列表
type ResourceMenu struct {
	Previous *ResourceMenu
	Items    []*Resource
}

// ShowPanel 展示列表
func (rm *ResourceMenu) ShowPanel() error {
	next, err := rm.showPanel()
	if err != nil {
		return err
	}

	if next != nil {
		return next.ShowPanel()
	}

	return nil
}

func (rm *ResourceMenu) showPanel() (*ResourceMenu, error) {
	template := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "➤ {{ .Name | white }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "➤ {{ .Name | green | cyan }}",
		Details: `
--------- Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Description }}
{{ if .Command }} {{- "Command:" | faint }}	{{ .Command }} {{ end }}
`,
	}

	var items []*Resource = rm.Items
	if rm.Previous != nil {
		items = append([]*Resource{goback2Prev}, rm.Items...)
	}

	// TODO：search rules
	searcher := func(input string, index int) bool {
		resource := items[index]
		name := strings.Replace(strings.ToLower(resource.Name), " ", "", -1)
		description := strings.Replace(strings.ToLower(resource.Description), " ", "", -1)
		command := strings.Replace(strings.ToLower(resource.Command), " ", "", -1)

		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input) || strings.Contains(description, input) || strings.Contains(command, input)
	}

	prompt := promptui.Select{
		Label:     "Pick your choice",
		Items:     items,
		Templates: template,
		Size:      5,
		Searcher:  searcher,
		//StartInSearchMode: true,
		HideSelected: true, // 选中后不打印显示
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run prompt: %v", err)
	}

	item := items[index]

	if item == goback2Prev {
		return rm.Previous, nil
	}

	if item.Children != nil {
		return &ResourceMenu{
			Previous: rm,
			Items:    item.Children,
		}, nil
	}

	if action, ok := registeredActions[item.SelectedAction]; ok && action != nil {
		return nil, action(item)
	}

	return nil, nil
}
