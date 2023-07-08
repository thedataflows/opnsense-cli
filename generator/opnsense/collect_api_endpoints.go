/**
*
* This file was converted from https://github.com/opnsense/docs/blob/master/collect_api_endpoints.py
*
**/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"

	git "github.com/go-git/go-git/v5"
	"github.com/thedataflows/go-commons/pkg/file"
)

const (
	excludeControllers = "Core/Api/FirmwareController.php"
	PostMethod         = "POST"
)

var (
	defaultBaseMethods = map[string][]map[string]string{
		"ApiMutableModelControllerBase": {
			{
				"command":    "set",
				"parameters": "",
				"method":     "POST",
			},
			{
				"command":    "get",
				"parameters": "",
				"method":     "GET",
			},
		},
		"ApiMutableServiceControllerBase": {
			{
				"command":    "status",
				"parameters": "",
				"method":     "GET",
			},
			{
				"command":    "start",
				"parameters": "",
				"method":     "POST",
			},
			{
				"command":    "stop",
				"parameters": "",
				"method":     "POST",
			},
			{
				"command":    "restart",
				"parameters": "",
				"method":     "POST",
			},
			{
				"command":    "reconfigure",
				"parameters": "",
				"method":     "POST",
			},
		},
	}
)

type Endpoint struct {
	Method        string
	Module        string
	Controller    string
	IsAbstract    bool
	BaseClass     string
	Command       string
	Parameters    []string
	Filename      string
	ModelFilename string
	Type          string
}

type Controller struct {
	Type       string
	Filename   string
	IsAbstract bool
	BaseClass  string
	Endpoints  []Endpoint
	Uses       []map[string]string
}

type TemplateData struct {
	Title          string
	TitleUnderline string
	Controllers    []Controller
}

func parseAPIPHP(srcFilename string) []Endpoint {
	baseFilename := filepath.Base(srcFilename)
	splitPath := strings.Split(srcFilename, "/")
	controller := strings.ToLower(strings.ReplaceAll(strings.Split(baseFilename, "Controller.php")[0], "(?<!^)(?=[A-Z])", "_"))
	moduleName := strings.ToLower(splitPath[len(splitPath)-3])

	data, err := os.ReadFile(srcFilename)
	if err != nil {
		log.Fatalf("Error reading file %v: %v", srcFilename, err)
	}
	dataStr := string(data)

	re := regexp.MustCompile(`\n([\w]*).*class.*Controller.*extends\s([\w|\\]*)`)
	m := re.FindAllStringSubmatch(dataStr, -1)
	baseClass := ""
	isAbstract := false
	if len(m) > 0 {
		s := strings.Split(m[0][2], "\\")
		baseClass = s[len(s)-1]
		isAbstract = m[0][1] == "abstract"
	}

	re = regexp.MustCompile(`\sprotected\sstatic\s\$internalModelClass\s=\s['|"]([\w|\\]*)['|"];`)
	if len(re.FindAllStringSubmatch(dataStr, -1)) == 0 {
		re = regexp.MustCompile(`\sprotected\sstatic\s\$internalServiceClass\s=\s['|"]([\w|\\]*)['|"];`)
	}
	modelFilename := ""
	m = re.FindAllStringSubmatch(dataStr, -1)
	if len(m) > 0 {
		m := re.FindAllStringSubmatch(dataStr, -1)[0][1]
		appLocation := strings.Join(splitPath[:len(splitPath)-5], "/")
		modelXML := fmt.Sprintf("%s/models/%s.xml", appLocation, strings.ReplaceAll(m, "\\", "/"))
		if _, err := os.Stat(modelXML); err == nil {
			modelFilename = strings.ReplaceAll(modelXML, "//", "/")
		}
	}

	re = regexp.MustCompile(`(\n\s+(private|public|protected)\s+function\s+(\w+)\((.*)\))`)
	functionCallouts := re.FindAllStringSubmatch(dataStr, -1)
	result := []Endpoint{}
	thisCommands := []string{}
	for idx, function := range functionCallouts {
		beginMarker := strings.Index(dataStr, functionCallouts[idx][0])
		// endMarker := -1
		endMarker := len(dataStr)
		if idx+1 < len(functionCallouts) {
			endMarker = strings.Index(dataStr, functionCallouts[idx+1][0])
		}
		codeBlock := dataStr[beginMarker+len(function[0]) : endMarker]
		if strings.HasSuffix(function[3], "Action") {
			thisCommands = append(thisCommands, function[3][:len(function[3])-6])
			parameters := strings.Split(strings.ReplaceAll(function[4], " ", ""), ",")
			if parameters[0] == "" {
				parameters = nil
			}
			record := Endpoint{
				Method:        "GET",
				Module:        moduleName,
				Controller:    controller,
				IsAbstract:    isAbstract,
				BaseClass:     baseClass,
				Command:       function[3][:len(function[3])-6],
				Parameters:    parameters,
				Filename:      baseFilename,
				ModelFilename: modelFilename,
			}
			if isAbstract {
				record.Type = "Abstract [non-callable]"
			} else if strings.Contains(controller, "service") {
				record.Type = "Service"
			} else {
				record.Type = "Resources"
			}
			switch {
			case strings.Contains(codeBlock, "request->isPost("):
				record.Method = PostMethod
			case strings.Contains(codeBlock, "$this->delBase"):
				record.Method = PostMethod
			case strings.Contains(codeBlock, "$this->addBase"):
				record.Method = PostMethod
			case strings.Contains(codeBlock, "$this->setBase"):
				record.Method = PostMethod
			case strings.Contains(codeBlock, "$this->toggleBase"):
				record.Method = PostMethod
			case strings.Contains(codeBlock, "$this->searchBase"):
				record.Method = "*"
			}
			result = append(result, record)
		}
	}
	if _, ok := defaultBaseMethods[baseClass]; ok {
		for _, item := range defaultBaseMethods[baseClass] {
			if !contains(thisCommands, item["command"]) {
				result = append(result, Endpoint{
					Type:          "Service",
					Method:        item["method"],
					Module:        moduleName,
					Controller:    controller,
					IsAbstract:    false,
					BaseClass:     baseClass,
					Command:       item["command"],
					Parameters:    nil,
					Filename:      baseFilename,
					ModelFilename: modelFilename,
				})
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Command < result[j].Command
	})

	return result
}

func sourceURL(repo string, srcFilename string) string {
	parts := strings.Split(srcFilename, "/")
	if repo == "plugins" {
		return fmt.Sprintf("https://github.com/opnsense/plugins/blob/master/%s", strings.Join(parts[findIndex(parts, "src")-2:], "/"))
	}
	return fmt.Sprintf("https://github.com/opnsense/core/blob/master/%s", strings.Join(parts[findIndex(parts, "src"):], "/"))
}

func findIndex(slice []string, val string) int {
	for i, item := range slice {
		if item == val {
			return i
		}
	}
	return -1
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// cloneGitRepo clones or pulls a git repository to a local directory
func cloneGitRepo(repoURL string, destinationDir string) {
	if len(destinationDir) == 0 {
		var err error
		// Set the repository URL and local directory
		repoURLSplit := strings.Split(repoURL, "/")
		destinationDir, err = filepath.Abs(strings.ReplaceAll(repoURLSplit[len(repoURLSplit)-1], ".git", ""))
		if err != nil {
			log.Fatalf("Error getting absolute path: %v", err)
		}
	}

	destinationDir = filepath.ToSlash(destinationDir)

	if file.IsDirectory(filepath.Join(destinationDir, ".git")) {
		// Open the repository
		r, err := git.PlainOpen(destinationDir)
		if err != nil {
			log.Fatalf("Error opening repository: %v", err)
		}

		// Get the working directory for the repository
		w, err := r.Worktree()
		if err != nil {
			log.Fatalf("Error getting work tree: %v", err)
		}

		log.Printf("Pulling changes from '%s' to '%s'", repoURL, destinationDir)
		// Pull the latest changes from the remote repository
		err = w.Pull(&git.PullOptions{
			// Depth:    1,
			Progress: os.Stdout,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			log.Fatalf("Error pulling changes: %v", err)
		}
		return
	}
	log.Printf("Cloning '%s' to '%s'", repoURL, destinationDir)
	// Clone the repository
	_, err := git.PlainClone(destinationDir,
		false,
		&git.CloneOptions{
			URL: repoURL,
			// Depth:    1,
			Progress: os.Stdout,
		})
	if err != nil {
		log.Fatalf("Error cloning repository: %v", err)
	}
}

func main() {
	sourceDir := flag.String("source", "", "source directory")
	repo := flag.String("repo", "core", "target repository")
	outputFile := flag.String("output", "", "output file")
	flag.Parse()

	if *sourceDir == "" {
		*sourceDir = filepath.Join(os.TempDir(), *repo)
		cloneGitRepo(fmt.Sprintf("https://github.com/opnsense/%s.git", *repo), *sourceDir)
	}

	allModules := map[string][][]Endpoint{}
	err := filepath.Walk(*sourceDir, func(path string, info os.FileInfo, err error) error {
		path = filepath.ToSlash(path)
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), "controller.php") || !strings.Contains(strings.ToLower(path), "mvc/app/controllers") || strings.Contains(strings.ToLower(path), excludeControllers) {
			return nil
		}
		payload := parseAPIPHP(path)
		if len(payload) > 0 {
			sPath := strings.Split(filepath.Dir(filepath.Dir(path)), "\\")
			moduleName := strings.ToLower(sPath[len(sPath)-1])
			if _, ok := allModules[moduleName]; !ok {
				allModules[moduleName] = [][]Endpoint{}
			}
			allModules[moduleName] = append(allModules[moduleName], payload)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking path: %v", err)
	}

	// Write to file
	if len(*outputFile) == 0 {
		*outputFile = fmt.Sprintf("%s.yaml", *repo)
	}
	log.Printf("Output file: %s", *outputFile)
	fOut, err := os.Create(filepath.Clean(*outputFile))
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer fOut.Close()

	fmt.Fprint(fOut, "commands:")

	// prepare the template
	templateFilename := fmt.Sprintf("%s.gotmpl", *outputFile)
	if _, err := os.Stat(templateFilename); err != nil {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			log.Fatal("No caller information")
		}
		templateFilename = filepath.Join(path.Dir(filename), "collect_api_endpoints.gotmpl")
	}
	log.Printf("Using template: %s", templateFilename)
	tmpl, err := template.ParseFiles(templateFilename)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	for moduleName, controllers := range allModules {
		templateData := TemplateData{
			Title:          strings.ToTitle(moduleName),
			TitleUnderline: strings.Repeat("~", len(moduleName)),
			Controllers:    []Controller{},
		}
		for _, controller := range controllers {
			payload := Controller{
				Type:       controller[0].Type,
				Filename:   controller[0].Filename,
				IsAbstract: controller[0].IsAbstract,
				BaseClass:  controller[0].BaseClass,
				Endpoints:  controller,
				Uses:       []map[string]string{},
			}
			if controller[0].ModelFilename != "" {
				payload.Uses = append(payload.Uses, map[string]string{
					"Type": "model",
					"Link": sourceURL(*repo, controller[0].ModelFilename),
					"Name": filepath.Base(controller[0].ModelFilename),
				})
			}
			templateData.Controllers = append(templateData.Controllers, payload)
		}

		// write template output
		err = tmpl.Execute(fOut, templateData)
		if err != nil {
			log.Fatalf("Error executing template: %v", err)
		}
	}
}
