package yamlResolver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

//YamlResolver resolves yaml files
type YamlResolver struct {
	data string
}

//LoadFile loads a file at the specified location and resolves all external references
func (o *YamlResolver) LoadFile(fileLoc string) error {
	dir := filepath.Dir(fileLoc)
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	fileName := filepath.Base(fileLoc)
	data, err := resolveYamlFile(abs, fileName, []string{})
	if err != nil {
		return err
	}
	o.data = data
	return nil
}

//String returns the resolved data as a single string
func (o *YamlResolver) String() string {
	return o.data
}

//Bytes returns the resolved data as bytes
func (o *YamlResolver) Bytes() []byte {
	return []byte(o.data)
}

//SaveFile saves the resolved data to the specified location
func (o *YamlResolver) SaveFile(fileLoc string) error {
	writer, err := os.Create(fileLoc)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = writer.WriteString(o.data)
	if err != nil {
		return err
	}
	return nil
}

func resolveYamlFile(workingDir, fileName string, heritage []string) (string, error) {
	fileLoc := filepath.Join(workingDir, fileName)
	if hasElem(heritage, fileLoc) {
		return "", errors.New("circular reference detected: " + fmt.Sprint(append(heritage, fileLoc)))
	}
	heritage = append(heritage, fileLoc)
	fileAsString, err := getYamlString(fileLoc)
	if err != nil {
		return "", err
	}
	newYamlSplit := []string{}
	split := strings.Split(fileAsString, "\n")
	for _, line := range split {
		processedLine, err := processLine(line, filepath.Dir(fileLoc), heritage)
		if err != nil {
			return "", err
		}
		newYamlSplit = append(newYamlSplit, processedLine)
	}
	return strings.Join(newYamlSplit, "\n"), nil
}

func processLine(line, dir string, heritage []string) (string, error) {
	split := strings.Split(line, "  ")
	for i, v := range split {
		prefix := ""
		prefixIndentadtion := 0
		if strings.HasPrefix(v, "- ") {
			v = strings.TrimPrefix(v, "- ")
			prefix = "- "
			prefixIndentadtion++
		}
		if strings.HasPrefix(v, "$ref") || strings.HasPrefix(v, "\"$ref\"") || strings.HasPrefix(v, "'$ref'") {
			fileLoc := strings.Split(v, " ")[1]
			fileLoc = cleanFileLoc(fileLoc)
			if !strings.HasPrefix(fileLoc, "#") {
				resolvedRef, err := resolveYamlFile(dir, cleanFileLoc(fileLoc), heritage)
				if err != nil {
					return "", err
				}
				resolvedRef = prefix + resolvedRef
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
		if strings.HasPrefix(lines[i], "- ") {
			tabSpace += "  "
		}
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

func hasElem(s interface{}, elem interface{}) bool {
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {
			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}
