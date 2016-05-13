package controller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/laincloud/registry-fake-pusher/rfp/model"
	"github.com/laincloud/registry-fake-pusher/rfp/utils/log"
)

// BlobController can download the blob from source location with sourceToken,
// then upload it to the target locaiton with targetToken.
type BlobController struct {
	source model.ImageLocation
	target model.ImageLocation

	sourceToken string
	targetToken string

	BlobSum string
	Content []byte
}

// NewBlobController will get the source an
func NewBlobController(s, t model.ImageLocation, b string, sJWT, tJWT string) (*BlobController, error) {
	bc := &BlobController{source: s, target: t, BlobSum: b}

	ac := NewAuthController("", "")
	if len(sJWT) > 0 {
		bc.sourceToken = sJWT
	} else {
		sToken, err := ac.GetAuthToken(bc.source.Registry, bc.source.Repository)
		if err != nil {
			return bc, err
		}
		bc.sourceToken = sToken
	}

	if len(tJWT) > 0 {
		bc.targetToken = tJWT
	} else {
		tToken, err := ac.GetAuthToken(bc.target.Registry, bc.target.Repository)
		if err != nil {
			return bc, err
		}
		bc.targetToken = tToken
	}

	return bc, nil
}

// Transfer download an blob from source repository,
// and upload it to target repository
func (bc *BlobController) Transfer() error {
	if err := bc.download(); err != nil {
		return err
	}

	if err := bc.upload(); err != nil {
		return err
	}
	return nil
}

func (bc *BlobController) download() error {
	log.Debugf("ready to download blob content")

	getBlobURL := fmt.Sprintf("%s/v2/%s/blobs/%s", bc.source.Registry,
		bc.source.Repository, bc.BlobSum)
	req, err := http.NewRequest("GET", getBlobURL, nil)
	if err != nil {
		return err
	}
	bc.addAuthHeader(req, true)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if bc.Content, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	log.Debugf("finish download blob content from %s", getBlobURL)
	return nil
}

func (bc *BlobController) upload() error {
	log.Debugf("ready to upload blob content")
	client := http.DefaultClient

	// initial upload
	initURL := fmt.Sprintf("%s/v2/%s/blobs/uploads/", bc.target.Registry,
		bc.target.Repository)
	initReq, err := http.NewRequest("POST", initURL, nil)
	if err != nil {
		return err
	}
	bc.addAuthHeader(initReq, false)
	initResp, err := client.Do(initReq)
	if err != nil {
		return err
	}
	if initResp.StatusCode > 300 {
		return fmt.Errorf("error when initial upload of blob: %s, status_code=%v",
			initURL, initResp.StatusCode)
	}
	location := initResp.Header.Get("Location")

	// upload blob
	uploadBlobURL := fmt.Sprintf("%s&digest=%s", location, bc.BlobSum)
	uploadBody := ioutil.NopCloser(strings.NewReader(string(bc.Content)))
	uploadReq, err := http.NewRequest("PUT", uploadBlobURL, uploadBody)
	if err != nil {
		return err
	}
	bc.addAuthHeader(uploadReq, false)
	uploadReq.Header.Set("Content-Length", string(len(bc.Content)))
	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return err
	}
	if uploadResp.StatusCode > 300 {
		return fmt.Errorf("error when initial upload of blob: %s, status_code=%v",
			uploadBlobURL, uploadResp.StatusCode)
	}

	log.Debugf("finish upload blob content to %s", initURL)
	return nil
}

func (bc *BlobController) addAuthHeader(req *http.Request, isSource bool) {
	if isSource {
		req.Header.Set("Authorization", "Bearer "+bc.sourceToken)
	} else {
		req.Header.Set("Authorization", "Bearer "+bc.targetToken)
	}
}
