package utils

import (
	"bufio"
	"encoding/json"
	"html/template"
	"os"

	"github.com/golang/glog"
)

type Output struct {
	Json         bool
	TemplatePath string
}

func (out Output) WriteOutput(diff interface{}) error {
	if out.Json {
		err := JSONify(diff)
		return err
	}
	return templateOutput(diff, out.TemplatePath)
}

var Templates = map[string]string{
	"single": "utils/output_templates/singleVersionOutput.txt",
	"multi":  "utils/output_templates/multiVersionOutput.txt",
	"hist":   "utils/output_templates/historyOutput.txt",
	"fs":     "utils/output_templates/fsOutput.txt",
}

func JSONify(diff interface{}) error {
	diffBytes, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		return err
	}
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	f.Write(diffBytes)
	return nil
}

// func getTemplatePath(diff interface{}) (string, error) {
// 	diffType := reflect.TypeOf(diff).String()
// 	fmt.Println(diffType)
// 	if path, ok := templates[diffType]; ok {
// 		return path, nil
// 	}
// 	return "", fmt.Errorf("No available template")
// }

func templateOutput(diff interface{}, tempPath string) error {
	// tempPath, err := getTemplatePath(diff)
	// if err != nil {
	// 	glog.Error(err)
	// }
	tmpl, err := template.ParseFiles(tempPath)
	if err != nil {
		glog.Error(err)
		return err
	}
	err = tmpl.Execute(os.Stdout, diff)
	if err != nil {
		glog.Error(err)
		return err
	}
	return nil
}
