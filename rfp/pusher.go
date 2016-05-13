package rfp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/laincloud/registry-fake-pusher/rfp/controller"
	"github.com/laincloud/registry-fake-pusher/rfp/model"
)

type RegistryFakePusher struct {
	SrcRegistry      string
	SrcRepository    string
	SrcTag           string
	TargetRegistry   string
	TargetRepository string
	TargetTag        string
	NewTag           string
}

func NewRegistryFakePusher(sReg, sRep, sTag, tReg, tRep, tTag, nTag string) (*RegistryFakePusher, error) {
	rfp := &RegistryFakePusher{
		SrcRegistry:      sReg,
		SrcRepository:    sRep,
		SrcTag:           sTag,
		TargetRegistry:   tReg,
		TargetRepository: tRep,
		TargetTag:        tTag,
		NewTag:           nTag}

	err := rfp.ValidRegistry()
	if err != nil {
		return nil, err
	}
	return rfp, nil
}

func (r *RegistryFakePusher) ValidRegistry() error {

	sReg, err := r.addProperScheme(r.SrcRegistry)
	if err != nil {
		return err
	}
	r.SrcRegistry = sReg

	tReg, err := r.addProperScheme(r.TargetRegistry)
	if err != nil {
		return err
	}
	r.TargetRegistry = tReg

	return nil
}

func (r *RegistryFakePusher) addProperScheme(reg string) (string, error) {
	if strings.HasPrefix(reg, "http") || strings.HasPrefix(reg, "https") {
		if err := r.ping(reg); err != nil {
			return "", err
		}
	} else {
		httpsReg := fmt.Sprintf("https://%s", reg)
		if err := r.ping(httpsReg); err == nil {
			return httpsReg, nil
		}

		httpReg := fmt.Sprintf("http://%s", reg)
		if err := r.ping(httpReg); err == nil {
			return httpReg, nil
		}
		return "", fmt.Errorf("registry %s is not accessable, please check again.", reg)
	}
	return reg, nil
}

func (r *RegistryFakePusher) ping(registry string) error {
	client := http.DefaultClient
	req, err := http.NewRequest("GET", registry, nil)
	if err != nil {
		return err
	}

	if _, err = client.Do(req); err != nil {
		return err
	}
	return nil
}

// FakePush gets the source and target manifest from the specify location,
// reconstructs a new image layer according the two manifests,
// overlays the new image layer to the target manifest generating a new manifest,
// and pushes the related blob and manifest to the registry
func (r *RegistryFakePusher) FakePush(srcJWT, targetJWT string, srcLayerCount int) error {

	sLoc := model.NewImageLocation(r.SrcRegistry, r.SrcRepository, r.SrcTag)
	tLoc := model.NewImageLocation(r.TargetRegistry, r.TargetRepository, r.TargetTag)

	sMc, err := controller.NewManifestController(sLoc, srcJWT)
	if err != nil {
		return fmt.Errorf("error create ManifestController for source manifest: %s", err)
	}
	tMc, err := controller.NewManifestController(tLoc, targetJWT)
	if err != nil {
		return fmt.Errorf("error create ManifestController for target manifest: %s", err)
	}

	var sIl, tIl model.ImageLayer
	for i := 0; i < srcLayerCount; i++ {
		if sIl, err = model.NewImageLayer(&model.Manifest{sMc.Manifest}, i); err != nil {
			return err
		}
		if tIl, err = model.NewImageLayer(&model.Manifest{tMc.Manifest}, 0); err != nil {
			return err
		}
		ic := controller.NewImageLayerController()
		newImageLayer, err := ic.GetToOverlayImageLayer(sIl, tIl)
		if err != nil {
			return fmt.Errorf("error get to overlay ImageLayer: %s", err)
		}
		tMc.Overlay(&newImageLayer, r.NewTag)

		if r.SrcRegistry != r.TargetRegistry || r.SrcRepository != r.TargetRepository {
			bc, err := controller.NewBlobController(sLoc, tLoc, sIl.FSLayer.BlobSum.String(), srcJWT, targetJWT)
			if err != nil {
				return fmt.Errorf("error get the blob controller : %s", err)
			}

			if err := bc.Transfer(); err != nil {
				return fmt.Errorf("error transter blob: %s", err)
			}
		}
	}
	tMc.Sign()

	if err := tMc.Push(); err != nil {
		return fmt.Errorf("error push new manifest : %s", err)
	}

	return nil
}
