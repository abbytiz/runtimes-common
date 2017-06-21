package differs

import (
	"bytes"
	"fmt"
	"os"
	"testing/runtimes-common/iDiff/utils"
)

// History compares the Docker history for each image.
func Package(d1file, d2file string) string {
	d1, e := utils.GetDirectory(d1file)
	if e != nil {
		panic(e)
		os.Exit(1)
	}
	d2, e := utils.GetDirectory(d2file)
	if e != nil {
		panic(e)
		os.Exit(1)
	}

	d1name := d1.Name
	d2name := d2.Name

	adds, dels, mods := utils.DiffDirectory(d1, d2)
	
	var buffer bytes.Buffer
	if adds == nil {
		buffer.WriteString("No files to diff\n")
	} else {
		s := fmt.Sprintf("These files have been added to %s\n", d1name)
		buffer.WriteString(s)
		if len(adds) == 0 {
			buffer.WriteString("none\n")
		}else {
			for _, f := range adds {
				s = fmt.Sprintf("%s\n", f)
				buffer.WriteString(s)
			}
		}

		s = fmt.Sprintf("These files have been deleted from %s\n", d1name)
		buffer.WriteString(s)
		if len(dels) == 0 {
			buffer.WriteString("none\n")
		}else {
			for _, f := range dels {
				s = fmt.Sprintf("%s\n", f)
				buffer.WriteString(s)
			}
		}
		s = fmt.Sprintf("These files have been changed between %s and %s\n", d1name, d2name)
		buffer.WriteString(s)
		if len(mods) == 0 {
			buffer.WriteString("none\n")
		}else {
			for _, f := range mods {
				s = fmt.Sprintf("%s\n", f)
				buffer.WriteString(s)
			}
		}	
	}
	return buffer.String()
}
