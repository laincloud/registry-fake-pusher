package controller

import (
	"encoding/json"
	"fmt"

	"github.com/laincloud/registry-fake-pusher/rfp/model"
	"github.com/laincloud/registry-fake-pusher/rfp/utils"
)

// ImageLayerController controls the construction of image
// layer, include basic msgs, environs and so on.
type ImageLayerController struct {
}

func NewImageLayerController() ImageLayerController {
	return ImageLayerController{}
}

// GetToOverlayImageLayer parses the src and target image layer, generates the newImageLayer
// used to overlay the target manifest.
func (ic ImageLayerController) GetToOverlayImageLayer(src, target model.ImageLayer) (newImageLayer model.ImageLayer, err error) {

	svc, err := utils.NewImgJSON([]byte(src.V1Compatibility))
	if err != nil {
		return newImageLayer, fmt.Errorf("error parse V1Comparivility from source ImageLayer: %s", err)
	}

	tvc, err := utils.NewImgJSON([]byte(target.V1Compatibility))
	if err != nil {
		return newImageLayer, fmt.Errorf("error parse V1Comparivility from target ImageLayer: %s", err)
	}

	svc.ID = utils.GenerateRandomID()
	svc.Parent = tvc.ID
	svc.ContainerConfig.Image = tvc.ID
	svc.Config = &(tvc.ContainerConfig)

	// TODO: add environment or something else in the config for the image
	// svc.Config.Env = append(svc.Config.Env, "LAIN_DOMAIN=lain")

	jsonData, err := json.Marshal(svc)
	if err != nil {
		return newImageLayer, fmt.Errorf("error marshal compatibility of new image layer: %s", err)
	}

	newImageLayer.FSLayer = src.FSLayer
	newImageLayer.V1Compatibility = string(jsonData)

	return newImageLayer, nil
}
