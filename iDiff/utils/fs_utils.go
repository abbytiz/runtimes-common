package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pmezard/go-difflib/difflib"
)

type Directory struct {
	Name string
	Files []string
	Dirs []Directory
}

func GetDirectory(name string) Directory {
	dirfile, e := ioutil.ReadFile(name)
	if e != nil {
		panic(e)
		os.Exit(1)
	}

	var dir Directory
	e = json.Unmarshal(dirfile, &dir)
	if e != nil {
		panic(e)
		os.Exit(1)
	}
	return dir
}

// Modification of difflib's unified differ 
func GetAddsAndDels(a, b []string, groups [][]difflib.OpCode) ([]string, []string) {
	var adds, dels []string
	for _, g := range groups {
		for _, c := range g {
			i1, i2, j1, j2 := c.I1, c.I2, c.J1, c.J2
			if c.Tag == 'r' || c.Tag == 'd' {
				for _, line := range a[i1:i2] {
					dels = append(dels, line)
				}
			}
			if c.Tag == 'r' || c.Tag == 'i' {
				for _, line := range b[j1:j2] {
					adds = append(adds, line)
				}
			}
		}
	}
	return adds, dels
}


func GetMatchStrings(a []string, matches []difflib.Match) []string {
	var matchstrings []string
	for i, m := range matches {
		if i != len(matches) - 1 {
			for _, line := range a[m.A : m.A + m.Size] {
				matchstrings = append(matchstrings, line)				
			}		
		}
	}
	return matchstrings
}

func GetModifiedFiles(path1, path2 string, files []string) []string {
	var mods []string
	for _, f := range files {
		f1path := fmt.Sprintf("%s%s", path1, f)
		f2path := fmt.Sprintf("%s%s", path2, f)
		if !CheckSameFile(f1path, f2path) {
			mods = append(mods, f)
		}	
	}
	return mods
}


func CompareFileEntries(d1, d2 Directory) ([]string, []string, []string) {
	e1 := d1.Files
	e2 := d2.Files
	matcher := difflib.NewMatcher(e1, e2)
	matchindexes := matcher.GetMatchingBlocks()
	diffindexes := matcher.GetGroupedOpCodes(0)

	matches := GetMatchStrings(e1, matchindexes)
	mods := GetModifiedFiles(d1.Name, d2.Name, matches)

	adds, dels := GetAddsAndDels(e1, e2, diffindexes)
	
	return adds, dels, mods
}

func CheckSameFile(f1name, f2name string) bool {
	// Check first if files differ in size and immediately return
	f1stat, err := os.Stat(f1name)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	f2stat, err := os.Stat(f2name)
	if err != nil {
		panic(err)
		os.Exit(1)
	}

	if f1stat.Size() != f2stat.Size() {
		return false	
	}
	
	// Next, check file contents
	f1, err := ioutil.ReadFile(f1name)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	f2, err := ioutil.ReadFile(f2name)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	if !bytes.Equal(f1, f2) {
		return false
	}
	return true
}


func DiffDirectory(d1, d2 Directory) ([]string, []string, []string) {
	adds, dels, mods := CompareFileEntries(d1, d2)
	return adds, dels, mods
}
