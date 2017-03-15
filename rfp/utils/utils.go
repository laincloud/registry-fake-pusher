package utils

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/manifest"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/runconfig"
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

func PatchLayer(svc *image.Image, tvc *image.Image) error {
	svc.ID = GenerateRandomID()
	svc.Parent = tvc.ID
	MergeConfig(&svc.ContainerConfig, tvc.Config)
	if svc.Config == nil {
		svc.Config = &runconfig.Config{}
	}
	MergeConfig(svc.Config, &svc.ContainerConfig)
	return nil
}

func MergeConfig(sConfig *runconfig.Config, tConfig *runconfig.Config) {
	if sConfig == nil || tConfig == nil {
		return
	}

	if sConfig.User == "" {
		sConfig.User = tConfig.User
	}
	if sConfig.Cmd.Len() == 0 {
		sConfig.Cmd = tConfig.Cmd
	}
	if sConfig.Entrypoint == nil {
		sConfig.Entrypoint = tConfig.Entrypoint
	}
	if sConfig.WorkingDir == "" {
		sConfig.WorkingDir = tConfig.WorkingDir
	}

	if len(sConfig.Env) == 0 {
		sConfig.Env = tConfig.Env
	} else {
		for _, imageEnv := range tConfig.Env {
			found := false
			imageEnvKey := strings.Split(imageEnv, "=")[0]
			for _, userEnv := range sConfig.Env {
				userEnvKey := strings.Split(userEnv, "=")[0]
				if imageEnvKey == userEnvKey {
					found = true
					break
				}
			}
			if !found {
				sConfig.Env = append(sConfig.Env, imageEnv)
			}
		}
	}

	if sConfig.Labels == nil {
		sConfig.Labels = map[string]string{}
	}
	for l, v := range tConfig.Labels {
		if _, ok := sConfig.Labels[l]; !ok {
			sConfig.Labels[l] = v
		}
	}

	if len(sConfig.Volumes) == 0 {
		sConfig.Volumes = tConfig.Volumes
	} else {
		for k, v := range tConfig.Volumes {
			sConfig.Volumes[k] = v
		}
	}

}
