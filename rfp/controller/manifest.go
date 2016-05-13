package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/laincloud/registry-fake-pusher/rfp/model"
	"github.com/laincloud/registry-fake-pusher/rfp/utils"
	"github.com/laincloud/registry-fake-pusher/rfp/utils/log"
)

// ManifestController controls the download and upload of
// the manifest, and the overlay to the manifest
type ManifestController struct {

	// ImageLocation is the location of the manifest
	model.ImageLocation

	// SignedManifest is the manifest detail msg
	model.SignedManifest

	// Token is the certificate to use registry api
	Token string
}

func NewManifestController(i model.ImageLocation, jwt string) (*ManifestController, error) {
	log.Debugf("new manifest controller for %s/%s:%s", i.Registry, i.Repository, i.Tag)

	mc := &ManifestController{ImageLocation: i}

	if len(jwt) > 0 {
		mc.Token = jwt
	} else {
		ac := NewAuthController("", "")
		token, err := ac.GetAuthToken(mc.ImageLocation.Registry, mc.ImageLocation.Repository)
		if err != nil {
			return mc, err
		}
		mc.Token = token
	}

	if err := mc.load(); err != nil {
		return mc, err
	}

	return mc, nil
}

func (mc *ManifestController) load() error {
	log.Debugf("ready to load manifest")

	url := mc.ImageLocation.GetManifestUrl()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	mc.addAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 {
		return fmt.Errorf("error when loading manifest from %s, status_code=%v",
			url, resp.StatusCode)
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(respBytes, &mc.Manifest)
	log.Debugf("finish load manifest from %s", url)
	return nil
}

func (mc *ManifestController) Sign() error {
	trustKey, err := utils.CreateTrustKey()
	if err != nil {
		return err
	}

	signed, err := utils.Sign(&mc.Manifest, trustKey)
	if err != nil {
		return err
	}

	mc.SignedManifest = model.SignedManifest{*signed}
	return nil
}

// Push pushes the new SignedMainfest in the ManifestController
// into the location specified by ImageLocation
func (mc *ManifestController) Push() error {
	log.Debugf("ready to push new manifest")

	mUploadURL := mc.ImageLocation.GetManifestUrl()
	req, err := http.NewRequest("PUT", mUploadURL, bytes.NewReader(mc.Raw))
	if err != nil {
		return err
	}
	mc.addAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 300 {
		return fmt.Errorf("error when pushing manifest from %s, status_code=%v",
			mUploadURL, resp.StatusCode)
	}

	log.Debugf("push new manifest success to %s", mUploadURL)
	return nil
}

// Overlay add a new ImageLayer i into the manifest,
// update the original manifest with Tag tag.
func (mc *ManifestController) Overlay(i *model.ImageLayer, tag string) {
	log.Debugf("ready to overlay image layer")

	m := &mc.Manifest
	mc.updateTag(tag)

	lastHistory := m.History[len(m.History)-1]
	lastLayer := m.FSLayers[len(m.FSLayers)-1]
	m.FSLayers = append(m.FSLayers, lastLayer)
	m.History = append(m.History, lastHistory)

	for i := len(m.History) - 1; i >= 1; i-- {
		m.FSLayers[i] = m.FSLayers[i-1]
		m.History[i] = m.History[i-1]
	}

	m.FSLayers[0] = i.FSLayer
	m.History[0] = i.History

	log.Debugf("finish overlay image layer")
}

func (mc *ManifestController) updateTag(tag string) {
	mc.ImageLocation.Tag = tag
	mc.Manifest.Tag = tag
}

func (mc *ManifestController) addAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+mc.Token)
}
