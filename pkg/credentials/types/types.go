package types

type Credentials struct {
	// APIToken is the API token used to authenticate with the Replicated API
	APIToken string `json:"token"`

	IsEnv        bool `json:"-"`
	IsConfigFile bool `json:"-"`
}
