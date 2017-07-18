package utils

import (
	"bufio"
	"encoding/json"
	"html/template"
	"os"

	"github.com/golang/glog"
)

// Output stores the information necessary to output a given diff
type Output struct {
	Json         bool
	TemplatePath string
}

// WriteOutput writes either the json or human readable format to Stdout
func (out Output) WriteOutput(diff interface{}) error {
	if out.Json {
		err := jsonify(diff)
		return err
	}
	return templateOutput(diff, out.TemplatePath)
}

// Templates stores paths to the template files for different diff outputs
var Templates = map[string]string{
	"single": "utils/output_templates/singleVersionOutput.txt",
	"multi":  "utils/output_templates/multiVersionOutput.txt",
	"hist":   "utils/output_templates/historyOutput.txt",
	"fs":     "utils/output_templates/fsOutput.txt",
}

func jsonify(diff interface{}) error {
	diffBytes, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		return err
	}
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()
	f.Write(diffBytes)
	return nil
}

func templateOutput(diff interface{}, tempPath string) error {
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
