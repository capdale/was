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
			"default",
			"suite/default_config.yaml",
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
