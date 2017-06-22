package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Directory struct {
	Root string
	Entries []string
}

func GetDirectory(dirpath string) (Directory, error) {
	dirfile, err := ioutil.ReadFile(dirpath)
	if err != nil {
		return Directory{}, err
	}

	var dir Directory
	err = json.Unmarshal(dirfile, &dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

// Checks for content differences between files of the same name from different directories
func getModifiedFiles(d1, d2 Directory) []string {
	d1files := d1.Entries
	d2files := d2.Entries

	if len(d1files) == 0 && len(d2files) == 0 {
		return nil
	}

	filematches := GetMatches(d1files, d2files)

	modified := []string{}
	for _, f := range filematches {
		f1path := fmt.Sprintf("%s%s", d1.Root, f)
		f2path := fmt.Sprintf("%s%s", d2.Root, f)
		
		f1stat, err := os.Stat(f1path)
		if err != nil {
			fmt.Printf("Error checking directory entry %s: %s\n", f, err)
			os.Exit(1)		
		}
		if !f1stat.IsDir() {
			same, err := checkSameFile(f1path, f2path)
			if err != nil {
				fmt.Printf("Error diffing contents of %s and %s: %s\n", f1path, f2path, err)
				os.Exit(1)
			}
			if !same {
				modified = append(modified, f)
			}					
		}	
	}
	return modified
}

func getAddedFiles(d1, d2 Directory) []string {
	if len(d1.Entries) == 0 && len(d2.Entries) == 0 {
		return nil
	}
	return GetAdditions(d1.Entries, d2.Entries)
}

func getDeletedFiles(d1, d2 Directory) []string {
	if len(d1.Entries) == 0 && len(d2.Entries) == 0 {
		return nil
	}
	return GetDeletions(d1.Entries, d2.Entries)
}

func compareFileEntries(d1, d2 Directory) ([]string, []string, []string) {
	adds := getAddedFiles(d1, d2)
	dels := getDeletedFiles(d1, d2)
	mods := getModifiedFiles(d1, d2)

	return adds, dels, mods
}

func checkSameFile(f1name, f2name string) (bool, error) {
	// Check first if files differ in size and immediately return
	f1stat, err := os.Stat(f1name)
	if err != nil {
		return false, err
	}
	f2stat, err := os.Stat(f2name)
	if err != nil {
		return false, err
	}

	if f1stat.Size() != f2stat.Size() {
		return false, nil
	}

	// Next, check file contents
	f1, err := ioutil.ReadFile(f1name)
	if err != nil {
		return false, err
	}
	f2, err := ioutil.ReadFile(f2name)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(f1, f2) {
		return false, nil
	}
	return true, nil
}

func DiffDirectory(d1, d2 Directory) ([]string, []string, []string) {
	// Diff file entries in the directories
	adds, dels, mods := compareFileEntries(d1, d2)
	return adds, dels, mods

	// TODO: Diff subdirectories within the directories
}
