package differs

import (
	"github.com/GoogleCloudPlatform/runtimes-common/iDiff/utils"
	"github.com/golang/glog"
)

// Package diffs two packages and compares their contents
func Package(dir1, dir2 string, json bool, eng bool) (string, error) {
	diff, err := getDiffOutput(dir1, dir2, json)
	if err != nil {
		return "", err
	}

	return diff, nil
}

func getDiffOutput(d1file, d2file string, json bool) (string, error) {
	d1, err := utils.GetDirectory(d1file)
	if err != nil {
		glog.Errorf("Error reading directory structure from file %s: %s\n", d1file, err)
		return "", err
	}
	d2, err := utils.GetDirectory(d2file)
	if err != nil {
		glog.Errorf("Error reading directory structure from file %s: %s\n", d2file, err)
		return "", err
	}

	dirDiff := utils.DiffDirectory(d1, d2)

	if json {
		return utils.JSONify(dirDiff)
	}

	err = utils.Output(dirDiff)
	return "", err
}
