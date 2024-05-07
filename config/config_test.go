package config_test

import (
	"testing"

	"github.com/capdale/was/config"
)

func TestConfig(t *testing.T) {
	cases := []struct {
		name string
		path string
	}{
		{
			"mock object",
			"suite/mock_email_config.yaml",
		},
		{
			"ses service",
			"suite/ses_email_config.yaml",
		},
		{
			"local storage",
			"suite/localstorage_config.yaml",
		},
		{
			"s3 service",
			"suite/s3_config.yaml",
		},
	}

	for _, tcase := range cases {
		t.Run(tcase.name, func(t *testing.T) {
			_, err := config.ParseConfig(tcase.path)
			if err != nil {
				t.Errorf("expected no error, but got %s", err.Error())
			}
		})
	}
}
