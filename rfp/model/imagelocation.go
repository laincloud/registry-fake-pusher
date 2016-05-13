package model

import (
	"fmt"
)

// ImageLocation describe the location msg for image
type ImageLocation struct {

	// Registry is the url of registry where the image located
	Registry string

	// Repository is the repository name of the image in Registry
	Repository string

	// Tag is the tag of the image in Registry
	Tag string
}

func NewImageLocation(reg, rep, tag string) ImageLocation {
	return ImageLocation{reg, rep, tag}
}

func (i ImageLocation) GetManifestUrl() string {
	url := fmt.Sprintf("%s/v2/%s/manifests/%s",
		i.Registry,
		i.Repository,
		i.Tag)

	return url
}
