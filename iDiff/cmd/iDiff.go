package cmd

import (
	"errors"
	"fmt"
	"os"
	"testing/runtimes-common/iDiff/differs"
	"testing/runtimes-common/iDiff/utils"

	"github.com/spf13/cobra"
)

// iDiff represents the iDiff command
var iDiffCmd = &cobra.Command{
	Use:   "iDiff [container1] [container2] [differ]",
	Short: "Compare two images.",
	Long:  `Compares two images using the specifed differ. `,
	Run: func(cmd *cobra.Command, args []string) {
		if valid, err := checkArgNum(args); !valid {
			fmt.Println(err)
			os.Exit(1)
		}
		// TODO: Use more effective mapping structure for differs
		// TODO: Logging errors and diff results instead of just printing
		if args[2] == "hist" {
			diff := differs.History(args[0], args[1])
			fmt.Println(diff)
		} else if args[2] == "dir" {
			diff, err := dirDiff(args[0], args[1])
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(diff)
		} else {
			fmt.Println("Unknown differ")
		}
	},
}

func dirDiff(img1, img2 string) (string, error) {
	dir1, path1, err := utils.ImageToDir(img1)
	if err != nil {
		return "", err
	}
	dir2, path2, err := utils.ImageToDir(img2)
	if err != nil {
		return "", err
	}
	diff := differs.Package(dir1, dir2)

	defer os.RemoveAll(path1)
	defer os.RemoveAll(path2)
	defer os.Remove(dir1)
	defer os.Remove(dir2)

	return diff, nil
}

func checkArgNum(args []string) (bool, error) {
	var err_message string
	if len(args) < 2 {
		err_message = "Please have at least two container IDs as arguments."
		return false, errors.New(err_message)
	} else if len(args) > 3 {
		err_message = "Too many arguments."
		return false, errors.New(err_message)
	} else {
		return true, nil
	}
}

func init() {
	RootCmd.AddCommand(iDiffCmd)
}
