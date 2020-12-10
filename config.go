package trisads

import "github.com/kelseyhightower/envconfig"

// Settings uses envconfig to load required settings from the environment and
// validate them in preparation for running the TRISA Directory Service.
type Settings struct {
	BindAddr        string `envconfig:"TRISADS_BIND_ADDR" default:":4433"`
	DatabaseDSN     string `envconfig:"TRISADS_DATABASE" required:"true"`
	SectigoUsername string `envconfig:"SECTIGO_USERNAME" required:"true"`
	SectigoPassword string `envconfig:"SECTIGO_PASSWORD" required:"true"`
	SendGridAPIKey  string `envconfig:"SENDGRID_API_KEY" required:"true"`
	ServiceEmail    string `envconfig:"TRISADS_SERVICE_EMAIL" default:"admin@vaspdirectory.net"`
	AdminEmail      string `envconfig:"TRISADS_ADMIN_EMAIL" default:"admin@trisa.io"`
}

// Config creates a new settings object, loading environment variables and defaults.
func Config() (_ *Settings, err error) {
	var conf Settings
	if err = envconfig.Process("trisads", &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
