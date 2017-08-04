package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containers/image/docker"
	"github.com/golang/glog"
)

func main() {
	// imageName := "hello-world"
	// imageName := "fedora"
	imageName := "gcr.io/gcp-runtimes/multi-base"
	// name := "multi-base"
	path := "finaldir"
	ref, err := docker.ParseReference("//" + imageName)
	if err != nil {
		panic(err)
	}

	img, err := ref.NewImage(nil)
	if err != nil {
		panic(err)
	}
	defer img.Close()

	imgSrc, err := ref.NewImageSource(nil, nil)
	if err != nil {
		panic(err)
	}

	if _, ok := os.Stat(path); ok != nil {
		os.MkdirAll(path, 0777)
	}

	for _, b := range img.LayerInfos() {
		bi, _, err := imgSrc.GetBlob(b)
		if err != nil {
			glog.Error(err)
		}
		gzf, err := gzip.NewReader(bi)
		if err != nil {
			glog.Error(err)
		}
		tr := tar.NewReader(gzf)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				glog.Fatalf(err.Error())
			}

			if strings.Contains(header.Name, ".wh.") || strings.Contains(header.Name, "/.wh.") {
				newName := strings.Replace(header.Name, ".wh.", "", 1)
				os.Remove(header.Name)
				os.RemoveAll(newName)
				continue
			}

			target := filepath.Join(path, header.Name)
			mode := header.FileInfo().Mode()
			switch header.Typeflag {

			// if its a dir and it doesn't exist create it
			case tar.TypeDir:
				if _, err := os.Stat(target); err != nil {
					if err := os.MkdirAll(target, mode); err != nil {
						return
					}
					continue
				}

			// if it's a file create it
			case tar.TypeReg:

				currFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
				if err != nil {
					return
				}
				defer currFile.Close()
				_, err = io.Copy(currFile, tr)
				if err != nil {
					return
				}
			}

		}

	}
}

type imageHistory struct {
	Created    time.Time `json:"created"`
	Author     string    `json:"author,omitempty"`
	CreatedBy  string    `json:"created_by,omitempty"`
	Comment    string    `json:"comment,omitempty"`
	EmptyLayer bool      `json:"empty_layer,omitempty"`
}

func writeToTar(tmpDir string) (string, error) {
	tarPath := tmpDir + ".tar"
	fw, err := os.Create(tarPath)
	tw := tar.NewWriter(fw)
	defer tw.Close()
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsDir() {
			return nil
		}
		newPath := path[len(tmpDir):]
		if len(newPath) == 0 {
			return nil
		}
		fr, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fr.Close()

		if h, err := tar.FileInfoHeader(info, newPath); err != nil {
			log.Fatalln(err)
		} else {
			h.Name = newPath
			if err = tw.WriteHeader(h); err != nil {
				log.Fatalln(err)
			}
		}
		if length, err := io.Copy(tw, fr); err != nil {
			log.Fatalln(err)
		} else {
			fmt.Println(length)
		}
		return nil
	}

	if err = filepath.Walk(tmpDir, walkFn); err != nil {
		return tarPath, err
	}
	return tarPath, nil
}
