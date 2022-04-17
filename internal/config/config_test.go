package config

import (
	"testing"

	"github.com/jon4hz/deadshot/internal/database"

	"github.com/stretchr/testify/assert"
)

func TestConfigInOsDir(t *testing.T) {
	assert.Nil(t, database.InitDB())
	load(true)
}
