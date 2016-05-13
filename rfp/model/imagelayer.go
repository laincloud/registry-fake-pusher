package model

import (
	"fmt"
)

// ImageLayer is one layer of an image
type ImageLayer struct {
	fsLayer

	history
}

func NewImageLayer(m *Manifest, index int) (ImageLayer, error) {
	var imageLayer ImageLayer
	if index >= len(m.FSLayers) {
		return imageLayer, fmt.Errorf("error getting imagelayer: index out of range")
	}
	imageLayer.FSLayer = m.FSLayers[index]
	imageLayer.History = m.History[index]
	return imageLayer, nil
}

// TODO: may used to specify the config of image layer
type ImageLayerConfig struct {
	Env []string
}
