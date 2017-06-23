package differs

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// AptDiff compares the packages installed by apt-get.
func AptDiff(img1, img2 string) string {
	return get_history_diff(img1, img2)
}

func getPackages(path string) (map[string]string, error) {
	packages := make(map[string]string)

	var layerStems []string

	layers, err := ioutil.ReadDir(path)
	if err != nil {
		return packages, err
	}
	for _, layer := range layers {
		layerStems = append(layerStems, filepath.Join(path, layer.Name(), "var/lib/dpkg/status"))
	}

	for _, statusFile := range layerStems {
		if _, err := os.Stat(statusFile); err == nil {

			if file, err := os.Open(statusFile); err == nil {
				// make sure it gets closed
				defer file.Close()

				var currPackage string
				// create a new scanner and read the file line by line
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := strings.Split(scanner.Text(), ": ")
					if len(line) == 2 {
						key := line[0]
						value := line[1]
						if key == "Package" {
							currPackage = value
						}
						if key == "Version" {
							packages[currPackage] = value
						}
					}

				}

			} else {
				return packages, err
			}
		} else {
			// status file does not exist in this layer
			continue
		}

	}
	return packages, nil
}
