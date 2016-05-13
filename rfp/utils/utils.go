package utils

import (
	"fmt"

	"github.com/docker/distribution/manifest"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/libtrust"
)

// CreateTrustKey generates a new trust-by-registry key using libtrust
func CreateTrustKey() (libtrust.PrivateKey, error) {
	trustKey, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		return nil, fmt.Errorf("error generating private key: %s", err)
	}
	return trustKey, nil
}

// Sign signs the manifest with the provided private key, returning a
// SignedManifest.
func Sign(m *manifest.Manifest, pk libtrust.PrivateKey) (*manifest.SignedManifest, error) {
	signed, err := manifest.Sign(m, pk)
	if err != nil {
		return nil, fmt.Errorf("error sign manifest : %s", err)
	}
	return signed, nil
}

// GenerateRandomID returns an unique id
func GenerateRandomID() string {
	return stringid.GenerateRandomID()
}

// Build an Image object from raw json data
func NewImgJSON(src []byte) (*image.Image, error) {
	return image.NewImgJSON(src)
}
