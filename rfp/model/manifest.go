package model

import (
	"github.com/docker/distribution/manifest"
)

type SignedManifest struct {
	manifest.SignedManifest
}

type Manifest struct {
	manifest.Manifest
}

type fsLayer struct {
	manifest.FSLayer
}

type history struct {
	manifest.History
}
