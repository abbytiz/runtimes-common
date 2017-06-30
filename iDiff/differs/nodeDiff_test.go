package differs

import (
	"reflect"
	"testing"

	"github.com/GoogleCloudPlatform/runtimes-common/iDiff/utils"
)

func TestGetNodePackages(t *testing.T) {
	testCases := []struct {
		descrip  string
		path     string
		expected map[string]utils.PackageInfo
		err      bool
	}{
		{
			descrip:  "no directory",
			path:     "testDirs/notThere",
			expected: map[string]utils.PackageInfo{},
			err:      true,
		},
		{
			descrip:  "no packages",
			path:     "testDirs/noPackages",
			expected: map[string]utils.PackageInfo{},
		},
		{
			descrip: "all packages in one layer",
			path:    "testDirs/packageOne",
			expected: map[string]utils.PackageInfo{
				"pac1": {Version: "1.0", Size: "4096"},
				"pac2": {Version: "2.0", Size: "4096"},
				"pac3": {Version: "3.0", Size: "4096"}},
		},
		{
			descrip: "many packages in different layers",
			path:    "testDirs/packageMany",
			expected: map[string]utils.PackageInfo{
				"pac1": {Version: "1.0", Size: "4096"},
				"pac2": {Version: "2.0", Size: "4096"},
				"pac3": {Version: "3.0", Size: "4096"},
				"pac4": {Version: "4.0", Size: "4096"},
				"pac5": {Version: "5.0", Size: "4096"}},
		},
	}

	for _, test := range testCases {
		packages, err := getNodePackages(test.path)
		if err != nil && !test.err {
			t.Errorf("Got unexpected error: %s", err)
		}
		if err == nil && test.err {
			t.Errorf("Expected error but got none.")
		}
		if !reflect.DeepEqual(packages, test.expected) {
			t.Errorf("Expected: %s but got: %s", test.expected, packages)
		}
	}
}
func TestReadPackageJSON(t *testing.T) {
	testCases := []struct {
		descrip  string
		path     string
		expected nodePackage
		err      bool
	}{
		{
			descrip: "Error on non-existent file",
			path:    "testDirs/not_real.json",
			err:     true,
		},
		{
			descrip:  "Parse JSON with exact fields",
			path:     "testDirs/exact.json",
			expected: nodePackage{"La-croix", "Lime"},
		},
		{
			descrip:  "Parse JSON with additional fields",
			path:     "testDirs/extra.json",
			expected: nodePackage{"La-croix", "Lime"},
		},
	}
	for _, test := range testCases {
		actual, err := readPackageJSON(test.path)
		if err != nil && !test.err {
			t.Errorf("Got unexpected error: %s", err)
		}
		if err == nil && test.err {
			t.Error("Expected errorbut got none.")
		}
		if !reflect.DeepEqual(actual, test.expected) {
			t.Errorf("Expected: %s but got: %s", test.expected, actual)
		}
	}
}
