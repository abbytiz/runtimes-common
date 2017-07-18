package differs

import (
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/runtimes-common/iDiff/utils"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// History compares the Docker history for each image.
func History(img1, img2 string, json bool, eng bool) (string, error) {
	return getHistoryDiff(img1, img2, json, eng)
}

func getHistoryList(img string, eng bool) ([]string, error) {
	validDocker, err := utils.ValidDockerVersion(eng)
	if err != nil {
		return []string{}, err
	}
	var history []image.HistoryResponseItem
	if validDocker {
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			return []string{}, err
		}
		history, err = cli.ImageHistory(ctx, img)
		if err != nil {
			return []string{}, err
		}
	} else {
		glog.Info("Docker version incompatible with api, shelling out to local Docker client.")
		history, err = utils.GetImageHistory(img)
		if err != nil {
			return []string{}, err
		}
	}

	strhistory := make([]string, len(history))
	for i, layer := range history {
		layerDescription := strings.TrimSpace(layer.CreatedBy)
		strhistory[i] = fmt.Sprintf("%s\n", layerDescription)
	}
	return strhistory, nil
}

type HistDiff struct {
	Image1 string
	Image2 string
	Adds   []string
	Dels   []string
}

func getHistoryDiff(image1, image2 string, json bool, eng bool) (string, error) {
	history1, err := getHistoryList(image1, eng)
	if err != nil {
		return "", err
	}
	history2, err := getHistoryList(image2, eng)
	if err != nil {
		return "", err
	}
	adds := utils.GetAdditions(history1, history2)
	dels := utils.GetDeletions(history1, history2)
	diff := HistDiff{image1, image2, adds, dels}
	if json {
		return utils.JSONify(diff)
	}
	err = utils.Output(diff)
	return "", err
}
