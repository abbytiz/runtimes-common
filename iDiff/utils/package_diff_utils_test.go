package utils

import (
	"reflect"
	"sort"
	"testing"
)

type ByPackage []Info

func (a ByPackage) Len() int           { return len(a) }
func (a ByPackage) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPackage) Less(i, j int) bool { return a[i].Package < a[j].Package }

type ByMultiPackage []MultiVersionInfo

func (a ByMultiPackage) Len() int           { return len(a) }
func (a ByMultiPackage) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMultiPackage) Less(i, j int) bool { return a[i].Package < a[j].Package }

type ByPackageInfo []PackageInfo

func (a ByPackageInfo) Len() int           { return len(a) }
func (a ByPackageInfo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPackageInfo) Less(i, j int) bool { return a[i].Version < a[j].Version }

func TestDiffMaps(t *testing.T) {
	testCases := []struct {
		descrip  string
		map1     interface{}
		map2     interface{}
		expected interface{}
	}{
		{
			descrip: "Missing Packages.",
			map1: map[string]PackageInfo{
				"pac1": {Version: "1.0", Size: "40"},
				"pac3": {Version: "3.0", Size: "60"}},
			map2: map[string]PackageInfo{
				"pac4": {Version: "4.0", Size: "70"},
				"pac5": {Version: "5.0", Size: "80"}},
			expected: PackageDiff{
				Packages1: map[string]PackageInfo{
					"pac1": {Version: "1.0", Size: "40"},
					"pac3": {Version: "3.0", Size: "60"}},
				Packages2: map[string]PackageInfo{
					"pac4": {Version: "4.0", Size: "70"},
					"pac5": {Version: "5.0", Size: "80"}},
				InfoDiff: []Info{}},
		},
		{
			descrip: "Different Versions and Sizes.",
			map1: map[string]PackageInfo{
				"pac2": {Version: "2.0", Size: "50"},
				"pac3": {Version: "3.0", Size: "60"}},
			map2: map[string]PackageInfo{
				"pac2": {Version: "2.0", Size: "45"},
				"pac3": {Version: "4.0", Size: "60"}},
			expected: PackageDiff{
				Packages1: map[string]PackageInfo{},
				Packages2: map[string]PackageInfo{},
				InfoDiff: []Info{
					{"pac2", PackageInfo{Version: "2.0", Size: "50"},
						PackageInfo{Version: "2.0", Size: "45"}},
					{"pac3", PackageInfo{Version: "3.0", Size: "60"},
						PackageInfo{Version: "4.0", Size: "60"}}},
			},
		},
		{
			descrip: "Identical packages, versions, and sizes",
			map1: map[string]PackageInfo{
				"pac1": {Version: "1.0", Size: "40"},
				"pac2": {Version: "2.0", Size: "50"},
				"pac3": {Version: "3.0", Size: "60"}},
			map2: map[string]PackageInfo{
				"pac1": {Version: "1.0", Size: "40"},
				"pac2": {Version: "2.0", Size: "50"},
				"pac3": {Version: "3.0", Size: "60"}},
			expected: PackageDiff{
				Packages1: map[string]PackageInfo{},
				Packages2: map[string]PackageInfo{},
				InfoDiff:  []Info{}},
		},
		{
			descrip: "MultiVersion call with identical Packages in different layers",
			map1: map[string]map[string]PackageInfo{
				"pac5": {"hash1/globalPath": {Version: "version", Size: "size"}},
				"pac3": {"hash1/notquite/localPath": {Version: "version", Size: "size"}},
				"pac4": {"samePlace": {Version: "version", Size: "size"}}},
			map2: map[string]map[string]PackageInfo{
				"pac5": {"hash2/globalPath": {Version: "version", Size: "size"}},
				"pac3": {"hash2/notquite/localPath": {Version: "version", Size: "size"}},
				"pac4": {"samePlace": {Version: "version", Size: "size"}}},
			expected: MultiVersionPackageDiff{
				Packages1: map[string]map[string]PackageInfo{},
				Packages2: map[string]map[string]PackageInfo{},
				InfoDiff:  []MultiVersionInfo{},
			},
		},
		{
			descrip: "MultiVersion Packages",
			map1: map[string]map[string]PackageInfo{
				"pac5": {"onlyImg1": {Version: "version", Size: "size"}},
				"pac4": {"hash1/samePlace": {Version: "version", Size: "size"}},
				"pac1": {"layer1/layer/node_modules/pac1": {Version: "1.0", Size: "40"}},
				"pac2": {"layer1/layer/usr/local/lib/node_modules/pac2": {Version: "2.0", Size: "50"},
					"layer2/layer/usr/local/lib/node_modules/pac2": {Version: "3.0", Size: "50"}}},
			map2: map[string]map[string]PackageInfo{
				"pac4": {"hash2/samePlace": {Version: "version", Size: "size"}},
				"pac1": {"layer2/layer/node_modules/pac1": {Version: "2.0", Size: "40"}},
				"pac2": {"layer3/layer/usr/local/lib/node_modules/pac2": {Version: "4.0", Size: "50"}},
				"pac3": {"layer2/layer/usr/local/lib/node_modules/pac2": {Version: "5.0", Size: "100"}}},
			expected: MultiVersionPackageDiff{
				Packages1: map[string]map[string]PackageInfo{
					"pac5": {"onlyImg1": {Version: "version", Size: "size"}},
				},
				Packages2: map[string]map[string]PackageInfo{
					"pac3": {"layer2/layer/usr/local/lib/node_modules/pac2": {Version: "5.0", Size: "100"}},
				},
				InfoDiff: []MultiVersionInfo{
					{
						Package: "pac1",
						Info1:   []PackageInfo{{Version: "1.0", Size: "40"}},
						Info2:   []PackageInfo{{Version: "2.0", Size: "40"}},
					},
					{
						Package: "pac2",
						Info1:   []PackageInfo{{Version: "2.0", Size: "50"}, {Version: "3.0", Size: "50"}},
						Info2:   []PackageInfo{{Version: "4.0", Size: "50"}},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		diff := diffMaps(test.map1, test.map2)
		diffVal := reflect.ValueOf(diff)
		testExpVal := reflect.ValueOf(test.expected)
		switch test.expected.(type) {
		case PackageDiff:
			expected := testExpVal.Interface().(PackageDiff)
			actual := diffVal.Interface().(PackageDiff)
			sort.Sort(ByPackage(expected.InfoDiff))
			sort.Sort(ByPackage(actual.InfoDiff))
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected Diff to be: %s but got:%s", expected, actual)
				return
			}
		case MultiVersionPackageDiff:
			expected := testExpVal.Interface().(MultiVersionPackageDiff)
			actual := diffVal.Interface().(MultiVersionPackageDiff)
			sort.Sort(ByMultiPackage(expected.InfoDiff))
			sort.Sort(ByMultiPackage(actual.InfoDiff))
			for _, pack := range expected.InfoDiff {
				sort.Sort(ByPackageInfo(pack.Info1))
				sort.Sort(ByPackageInfo(pack.Info2))
			}
			for _, pack2 := range actual.InfoDiff {
				sort.Sort(ByPackageInfo(pack2.Info1))
				sort.Sort(ByPackageInfo(pack2.Info2))
			}
			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected Diff to be: %s but got:%s", expected, actual)
				return
			}
		}
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		descrip     string
		VersionList []PackageInfo
		Layers      []string
		currLayer   string
		currVersion PackageInfo
		index       int
		ok          bool
	}{
		{
			descrip:     "Does contain",
			VersionList: []PackageInfo{{Version: "2", Size: "b"}, {Version: "1", Size: "a"}},
			Layers:      []string{"1/global", "2/local"},
			currLayer:   "3/local",
			currVersion: PackageInfo{Version: "1", Size: "a"},
			index:       1,
			ok:          true,
		},
		{
			descrip:     "Not contained",
			VersionList: []PackageInfo{{Version: "1", Size: "a"}, {Version: "2", Size: "b"}},
			Layers:      []string{"1/global", "2/local"},
			currLayer:   "3/global",
			currVersion: PackageInfo{Version: "2", Size: "a"},
			index:       0,
			ok:          false,
		},
		{
			descrip:     "Does contain but path doesn't match",
			VersionList: []PackageInfo{{Version: "1", Size: "a"}, {Version: "2", Size: "b"}},
			Layers:      []string{"1/local", "2/local"},
			currLayer:   "3/global",
			currVersion: PackageInfo{Version: "1", Size: "a"},
			index:       0,
			ok:          false,
		},
		{
			descrip: "Does contain but layer doesn't match",
			VersionList: []PackageInfo{{Version: "1", Size: "a", Layer: "L1"},
				{Version: "2", Size: "b", Layer: "L2"}},
			Layers:      []string{"1/local", "2/local"},
			currLayer:   "3/global",
			currVersion: PackageInfo{Version: "1", Size: "a"},
			index:       0,
			ok:          false,
		},
		{
			descrip:     "Layers and Versions not of same length",
			VersionList: []PackageInfo{{Version: "1", Size: "a"}, {Version: "2", Size: "b"}},
			Layers:      []string{"1/local"},
			currLayer:   "3/global",
			currVersion: PackageInfo{Version: "1", Size: "a"},
			index:       0,
			ok:          false,
		},
	}
	for _, test := range testCases {
		index, ok := contains(test.VersionList, test.Layers, test.currLayer, test.currVersion)
		if test.ok != ok {
			t.Errorf("Expected status: %t, but got: %t", test.ok, ok)
		}
		if test.index != index {
			t.Errorf("Expected index: %d, but got: %d", test.index, index)
		}
	}
}
func TestCheckPackageMapType(t *testing.T) {
	testCases := []struct {
		descrip       string
		map1          interface{}
		map2          interface{}
		expectedType  reflect.Type
		expectedMulti bool
		err           bool
	}{
		{
			descrip: "Map arguments not maps",
			map1:    "not a map",
			map2:    "not a map either",
			err:     true,
		},
		{
			descrip: "Map arguments not same type",
			map1:    map[string]int{},
			map2:    map[int]string{},
			err:     true,
		},
		{
			descrip:      "Single Version Package Maps",
			map1:         map[string]PackageInfo{},
			map2:         map[string]PackageInfo{},
			expectedType: reflect.TypeOf(map[string]PackageInfo{}),
		},
		{
			descrip:       "MultiVersion Package Maps",
			map1:          map[string]map[string]PackageInfo{},
			map2:          map[string]map[string]PackageInfo{},
			expectedType:  reflect.TypeOf(map[string]map[string]PackageInfo{}),
			expectedMulti: true,
		},
	}
	for _, test := range testCases {
		actualType, actualMulti, err := checkPackageMapType(test.map1, test.map2)
		if err != nil && !test.err {
			t.Errorf("Got unexpected error: %s", err)
		}
		if err == nil && test.err {
			t.Error("Expected error but got none.")
		}
		if actualType != test.expectedType {
			t.Errorf("Expected type: %s but got: %s", test.expectedType, actualType)
		}
		if actualMulti != test.expectedMulti {
			t.Errorf("Expected multi: %t but got %t", test.expectedMulti, actualMulti)
		}
	}
}
func TestBuildLayerTargets(t *testing.T) {
	testCases := []struct {
		descrip  string
		path     string
		target   string
		expected []string
		err      bool
	}{
		{
			descrip:  "Filter out non directories",
			path:     "testTars/la-croix1-actual",
			target:   "123",
			expected: []string{},
		},
		{
			descrip:  "Error on bad directory path",
			path:     "test_files/notReal",
			target:   "123",
			expected: []string{},
			err:      true,
		},
		{
			descrip:  "Filter out non-directories and get directories",
			path:     "testTars/la-croix3-full",
			target:   "123",
			expected: []string{"testTars/la-croix3-full/nest/123", "testTars/la-croix3-full/nested-dir/123"},
		},
	}
	for _, test := range testCases {
		layers, err := BuildLayerTargets(test.path, test.target)
		if err != nil && !test.err {
			t.Errorf("Got unexpected error: %s", err)
		}
		if err == nil && test.err {
			t.Errorf("Expected error but got none: %s", err)
		}
		sort.Strings(test.expected)
		sort.Strings(layers)
		if !reflect.DeepEqual(test.expected, layers) {
			t.Errorf("Expected: %s, but got: %s.", test.expected, layers)
		}
	}
}
