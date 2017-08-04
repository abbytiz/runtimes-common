package utils

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/containers/image/directory"
	"github.com/containers/image/docker"
	"github.com/containers/image/types"
	"github.com/golang/glog"
)

var sourceToPrepMap = map[string]Prepper{
	"ID":  IDPrepper{},
	"URL": CloudPrepper{},
	"tar": TarPrepper{},
}

var sourceCheckMap = map[string]func(string) bool{
	"ID":  CheckImageID,
	"URL": CheckImageURL,
	"tar": CheckTar,
}

type Image struct {
	Source  string
	FSPath  string
	History []string
	Layers  []string
}

type ImagePrepper struct {
	Source string
}

type Prepper interface {
	ImageToFS() (string, error)
}

func (p ImagePrepper) GetImage() (Image, error) {
	glog.Infof("Starting prep for image %s", p.Source)
	img := p.Source

	var prepper Prepper
	for source, check := range sourceCheckMap {
		if check(img) {
			typePrepper := reflect.TypeOf(sourceToPrepMap[source])
			prepper = reflect.New(typePrepper).Interface().(Prepper)
			reflect.ValueOf(prepper).Elem().Field(0).Set(reflect.ValueOf(p))
			break
		}
	}
	if prepper == nil {
		return Image{}, errors.New("Could not retrieve image from source")
	}

	imgPath, err := prepper.ImageToFS()
	if err != nil {
		return Image{}, err
	}

	history, err := getHistory(imgPath)
	if err != nil {
		return Image{}, err
	}

	glog.Infof("Finished prepping image %s", p.Source)
	return Image{
		Source:  img,
		FSPath:  imgPath,
		History: history,
	}, nil
}

type histJSON struct {
	History []histLayer `json:"history"`
}

type histLayer struct {
	Created    string `json:"created"`
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer"`
}

func getHistory(imgPath string) ([]string, error) {
	glog.Info("Obtaining image history")
	histList := []string{}
	contents, err := ioutil.ReadDir(imgPath)
	if err != nil {
		return histList, err
	}

	for _, item := range contents {
		if filepath.Ext(item.Name()) == ".json" && item.Name() != "manifest.json" {
			file, err := ioutil.ReadFile(filepath.Join(imgPath, item.Name()))
			if err != nil {
				return histList, err
			}
			var histJ histJSON
			json.Unmarshal(file, &histJ)
			if len(histList) != 0 {
				glog.Error("Multiple history sources detected for image at " + imgPath + ", history diff may be incorrect.")
				break
			}
			for _, layer := range histJ.History {
				histList = append(histList, layer.CreatedBy)
			}
		}
	}
	return histList, nil
}

func getImageFromTar(tarPath string) (string, error) {
	glog.Info("Extracting image tar to obtain image file system")
	err := ExtractTar(tarPath)
	if err != nil {
		return "", err
	}
	path := strings.TrimSuffix(tarPath, filepath.Ext(tarPath))
	return path, nil
}

// CloudPrepper prepares images sourced from a Cloud registry
type CloudPrepper struct {
	ImagePrepper
}

func pullAndSaveImage(image string) (string, error) {
	URLPattern := regexp.MustCompile("^.+/(.+(:.+){0,1})$")
	URLMatch := URLPattern.FindStringSubmatch(image)
	imageName := strings.Replace(URLMatch[1], ":", "", -1)
	// imageURL := strings.TrimSuffix(image, URLMatch[2])

	ref, err := docker.ParseReference("//" + image)
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

	tmpDir, err := ioutil.TempDir(".", imageName+"-")
	if err != nil {
		glog.Error(err)
	}
	tmpDirRef, err := directory.NewReference(tmpDir)
	if err != nil {
		glog.Error(err)
	}
	dest, err := tmpDirRef.NewImageDestination(nil)
	if err != nil {
		glog.Error(err)
	}

	defer func() {
		if err := dest.Close(); err != nil {
			glog.Error(err)
		}
	}()

	for _, b := range img.LayerInfos() {
		bi, blobSize, err := imgSrc.GetBlob(b)
		if err != nil {
			panic(err)
		}
		newLayerDir, err := ioutil.TempDir(tmpDir, "layer-")
		if err != nil {
			glog.Error(err)
		}
		newLayerRef, err := directory.NewReference(newLayerDir)
		if err != nil {
			glog.Error(err)
		}
		layerDest, err := newLayerRef.NewImageDestination(nil)
		if err != nil {
			glog.Error(err)
		}

		if _, err := layerDest.PutBlob(bi, types.BlobInfo{Digest: b.Digest, Size: blobSize}); err != nil {
			if closeErr := bi.Close(); closeErr != nil {
				glog.Error(closeErr)
			}
		}

	}

	manifest, _, err := img.Manifest()
	if err != nil {
		glog.Error(err)
	}
	dest.PutManifest(manifest)

	tarPath, err := writeToTar(tmpDir)
	if err != nil {
		glog.Error(err)
		return tarPath, err
	}
	return tarPath, nil
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

func (p CloudPrepper) ImageToFS() (string, error) {
	tarPath, err := pullAndSaveImage(p.Source)
	if err != nil {
		glog.Error(err)
		return tarPath, err
	}
	// defer os.Remove(tarPath)
	return getImageFromTar(tarPath)

	// layers, _ := ioutil.ReadDir(tmpDir)
	// for _, layerFolder := range layers {
	// 	layerFolderPath := filepath.Join(tmpDir, layerFolder.Name())
	// 	tars, _ := ioutil.ReadDir(layerFolderPath)
	// 	for _, layer := range tars {
	// 		path := filepath.Join(layerFolderPath, layer.Name())
	// 		fmt.Println("UNPACKING ", path)
	// 		target := strings.TrimSuffix(path, filepath.Ext(layer.Name()))
	// 		fmt.Println("path ", path, "\ntarget ", target)
	// 		UnTar(path, target)
	// 		defer os.Remove(path)
	// 	}

	// }
	// return tmpDir, nil
}

type IDPrepper struct {
	ImagePrepper
}

func (p IDPrepper) ImageToFS() (string, error) {
	// check client compatibility with Docker API
	valid, err := ValidDockerVersion()
	if err != nil {
		return "", err
	}
	var tarPath string
	if !valid {
		glog.Info("Docker version incompatible with api, shelling out to local Docker client.")
		tarPath, err = imageToTarCmd(p.Source, p.Source)
	} else {
		tarPath, err = saveImageToTar(p.Source, p.Source)
	}
	if err != nil {
		return "", err
	}

	defer os.Remove(tarPath)
	return getImageFromTar(tarPath)
}

type TarPrepper struct {
	ImagePrepper
}

func (p TarPrepper) ImageToFS() (string, error) {
	return getImageFromTar(p.Source)
}
