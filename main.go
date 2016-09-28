package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func main() {
	// fmt.Println(Resolve("./parent.yaml"))
	fmt.Println(Resolve("./swaggerSpec/index.yaml"))
}

func Resolve(fileLoc string) (string, error) {
	dir := filepath.Dir(fileLoc)
	abs, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}
	fileName := filepath.Base(fileLoc)
	return resolveYamlFile(abs, fileName)
}

func resolveYamlFile(workingDir, fileName string) (string, error) {
	fileLoc := filepath.Join(workingDir, fileName)
	fileAsString, err := getYamlString(fileLoc)
	if err != nil {
		return "", err
	}
	newYamlSplit := []string{}
	split := strings.Split(fileAsString, "\n")
	for _, line := range split {
		processedLine, err := processLine(line, filepath.Dir(fileLoc))
		if err != nil {
			return "", err
		}
		newYamlSplit = append(newYamlSplit, processedLine)
	}
	return strings.Join(newYamlSplit, "\n"), nil
}

func processLine(line, dir string) (string, error) {
	split := strings.Split(line, "  ")
	for i, v := range split {
		if strings.HasPrefix(v, "$ref") {
			fileLoc := strings.Split(v, " ")[1]
			fileLoc = cleanFileLoc(fileLoc)
			if !strings.HasPrefix(fileLoc, "#") {
				resolvedRef, err := resolveYamlFile(dir, cleanFileLoc(fileLoc))
				if err != nil {
					return "", err
				}
				return addIndentation(resolvedRef, i), nil
			}
		}
	}
	return strings.Join(split, "  "), nil
}

func addIndentation(s string, tabs int) string {
	tabSpace := ""
	for i := 0; i < tabs; i++ {
		tabSpace += "  "
	}

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = tabSpace + lines[i]
	}
	return strings.Join(lines, "\n")
}

func getYamlString(fileLoc string) (string, error) {
	fileLoc = cleanFileLoc(fileLoc)
	fileBytes, err := ioutil.ReadFile(fileLoc)
	if err != nil {
		return "", err
	}

	ret := string(fileBytes[:])
	ret = strings.TrimSuffix(ret, "\n")
	return ret, nil

}

func cleanFileLoc(fileLoc string) string {
	fileLoc = strings.Replace(fileLoc, "\"", "", -1)
	fileLoc = strings.Replace(fileLoc, "'", "", -1)
	return fileLoc
}
