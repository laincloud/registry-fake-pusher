package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/homedir"

	"github.com/laincloud/registry-fake-pusher/rfp/model"
	"github.com/laincloud/registry-fake-pusher/rfp/utils/log"
)

type AuthController struct {
	configDir      string
	configFileName string
}

// NewAuthController create an AuthController based on the passed dir and fileName.
func NewAuthController(dir, fileName string) *AuthController {
	ac := &AuthController{
		configDir:      dir,
		configFileName: fileName}

	if dir == "" {
		configDir := os.Getenv("DOCKER_CONFIG")
		ac.configDir = configDir
		if configDir == "" {
			ac.configDir = filepath.Join(homedir.Get(), ".docker")
		}
	}

	if fileName == "" {
		ac.configFileName = "config.json"
	}
	return ac
}

// GetAuthToken get the auth token for the specified registry and repository,
// if auth is not need, an empty string will be returned.
func (ac *AuthController) GetAuthToken(registry, repository string) (string, error) {
	params, err := ac.ping(registry)
	if err != nil {
		return "", nil
	}

	if params != nil {
		authConfigs, err := ac.loadConfig()
		if err != nil {
			return "", err
		}
		registry = ac.formatRegistry(registry)
		authConfig := authConfigs.AuthConfigs[registry]
		token, err := ac.authorize(repository, &authConfig, params)
		return token, err
	}

	return "", nil
}

func (ac *AuthController) formatRegistry(registry string) string {
	switch {
	case strings.HasPrefix(registry, "http://"):
		return registry[7:len(registry)]
	case strings.HasPrefix(registry, "https://"):
		return registry[8:len(registry)]
	default:
		return registry
	}
}

// ping will connect the registry, check whether needing authorization,
// if do not get the params for authorization, means no need for authorization.
func (ac *AuthController) ping(registry string) (map[string]string, error) {
	log.Debugf("ping registry : %s", registry)

	client := http.DefaultClient
	url := fmt.Sprintf("%s/v2/", registry)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ac.parseAuthHeader(resp.Header), nil
	}
	return nil, nil
}

func (ac *AuthController) parseAuthHeader(header http.Header) map[string]string {
	authParams := make(map[string]string)
	for _, h := range header[http.CanonicalHeaderKey("WWW-Authenticate")] {
		authParams = ac.parseValueAndParams(h)
	}
	return authParams
}

func (ac *AuthController) parseValueAndParams(header string) map[string]string {
	params := make(map[string]string)
	msgs := strings.Split(header, ",")
	for i := 0; i < len(msgs); i++ {
		values := strings.Split(msgs[i], "=")
		if len(values) > 0 {
			params[values[0]] = values[1]
		}
	}
	return params
}

// loadConfig reads the configuration files in the given directory, and sets up
// the auth config information and return values.
func (ac *AuthController) loadConfig() (*model.ConfigFile, error) {
	log.Debugf("load config file from %s/%s", ac.configDir, ac.configFileName)

	configFile := model.ConfigFile{
		AuthConfigs: make(map[string]model.AuthConfig),
		Filename:    filepath.Join(ac.configDir, ac.configFileName),
	}

	_, err := os.Stat(configFile.Filename)
	if err == nil {
		file, err := os.Open(configFile.Filename)
		if err != nil {
			return &configFile, err
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&configFile); err != nil {
			return &configFile, err
		}

		for addr, cfg := range configFile.AuthConfigs {
			cfg.Username, cfg.Password, err = ac.decodeAuth(cfg.Auth)
			if err != nil {
				return &configFile, err
			}
			cfg.Auth = ""
			cfg.ServerAddress = addr
			configFile.AuthConfigs[addr] = cfg
		}

		return &configFile, nil
	}
	return &configFile, err
}

// decodeAuth decodes a base64 encoded string and returns username and password.
func (ac *AuthController) decodeAuth(authStr string) (string, string, error) {
	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", fmt.Errorf("Something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("Invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")

	return arr[0], password, nil
}

func (ac *AuthController) authorize(repository string, authConfig *model.AuthConfig, params map[string]string) (string, error) {
	log.Debugf("get token for repository: %s", repository)

	client := http.DefaultClient
	url := params["Bearer realm"]
	url = url[1 : len(url)-1]
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	service := params["service"]
	reqParams := req.URL.Query()
	reqParams.Add("service", service[1:len(service)-1])
	reqParams.Add("scope", "repository:"+repository+":push,pull")
	reqParams.Add("account", authConfig.Username)
	req.SetBasicAuth(authConfig.Username, authConfig.Password)
	req.URL.RawQuery = reqParams.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	token := model.Token{}
	json.Unmarshal(respBytes, &token)
	return token.Token, nil
}
