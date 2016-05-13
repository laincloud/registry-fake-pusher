package model

// AuthConfig contains authorization information for connecting to a Registry
type AuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Email         string `json:"email"`
	Auth          string `json:"auth"`
	ServerAddress string `json:"serveraddress,omitempty"`
}

type Token struct {
	Token string `json:"token"`
}

// ConfigFile ~/.docker/config.json file info
type ConfigFile struct {
	AuthConfigs map[string]AuthConfig `json:"auths"`
	Filename    string                // Note: not serialized - for internal use only
}
