package config_test

import (
	"testing"

	"github.com/GoROSEN/rosen-apiserver/core/config"
)

func TestNewConfig(t *testing.T) {

	c := config.GetConfig()
	if c == nil {
		t.Error("cannot get config")
	}
}
