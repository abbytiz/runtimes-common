package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

func test() {
	ref, err := docker.ParseReference("//gcr.io/gcp-runtimes/multi-base")
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

	tmpDir, err := ioutil.TempDir(".", "layers-")
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
		fmt.Println(b.Digest)
		fmt.Println(b.URLs)
		bi, _, err := imgSrc.GetBlob(b)
		if err != nil {
			panic(err)
		}
		fmt.Println("Got blob: %s", bi)
		buf := new(bytes.Buffer)
		buf.ReadFrom(bi)
		newStr := buf.String()

		fmt.Println("the blob: ", newStr)

		// if _, err := dest.PutBlob(bi, types.BlobInfo{Digest: b.Digest, Size: blobSize}); err != nil {
		// 	if closeErr := bi.Close(); closeErr != nil {
		// 		glog.Error(closeErr)
		// 	}
		// }
	}

	manifest, s, err := img.Manifest()
	if err != nil {
		glog.Error(err)
	}
	dest.PutManifest(manifest)
	fmt.Println(s)

	if err != nil {
		panic(err)
	}
}

func (p CloudPrepper) ImageToFS() (string, error) {
	// imageName := "hello-world"
	// imageName := "fedora"
	// imageName := "gcr.io/gcp-runtimes/multi-base"
	name := "testImg"
	ref, err := docker.ParseReference("//" + p.Source)
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

	// manByte, str, err := imgSrc.GetManifest()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(str)

	tmpDir, err := ioutil.TempDir(".", name)
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
		// fmt.Println(b.Digest)
		// fmt.Println(b.URLs)
		bi, blobSize, err := imgSrc.GetBlob(b)
		if err != nil {
			panic(err)
		}
		// fmt.Println("Got blob: %s", bi)
		// buf := new(bytes.Buffer)
		// buf.ReadFrom(bi)
		// newStr := buf.String()
		// fmt.Println("the blob: ", newStr)
		newLayerDir, err := ioutil.TempDir(tmpDir, "layer")
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

	layers, _ := ioutil.ReadDir(tmpDir)
	for _, layerFolder := range layers {
		layerFolderPath := filepath.Join(tmpDir, layerFolder.Name())
		tars, _ := ioutil.ReadDir(layerFolderPath)
		for _, layer := range tars {
			path := filepath.Join(layerFolderPath, layer.Name())
			fmt.Println("UNPACKING ", path)
			target := strings.TrimSuffix(path, filepath.Ext(layer.Name()))
			fmt.Println("path ", path, "\ntarget ", target)
			UnTar(path, target)
			defer os.Remove(path)
		}

	}
	return tmpDir, nil
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
