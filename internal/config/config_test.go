package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTestDefaults ensures required env vars are set for non-demo apps.
func setTestDefaults(t *testing.T) {
	t.Helper()
	if os.Getenv("APP_NAME") == "" {
		os.Setenv("APP_NAME", "test-app")
		t.Cleanup(func() { os.Unsetenv("APP_NAME") })
	}
}

func TestGetConfig(t *testing.T) {
	ResetForTesting()
	setTestDefaults(t)

	os.Setenv("SERVER_LISTEN_PORT", "9090")
	defer os.Unsetenv("SERVER_LISTEN_PORT")

	config, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "9090", config.ServerPort)

	config2, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, config, config2)
}

func TestGetConfigDefaults(t *testing.T) {
	ResetForTesting()
	setTestDefaults(t)

	os.Unsetenv("SERVER_LISTEN_PORT")
	os.Unsetenv("DATABASE_URL")

	config, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "3000", config.ServerPort)
	assert.Equal(t, "sqlite:///db/app.db", config.DatabaseURL)
}

func TestMustGetConfig(t *testing.T) {
	ResetForTesting()
	setTestDefaults(t)

	os.Setenv("SERVER_LISTEN_PORT", "7070")
	defer os.Unsetenv("SERVER_LISTEN_PORT")

	config := MustGetConfig()
	assert.Equal(t, "7070", config.ServerPort)
}

func TestConfigEnvOverride(t *testing.T) {
	ResetForTesting()

	os.Setenv("SERVER_LISTEN_PORT", "5555")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("APP_NAME", "testapp")
	defer func() {
		os.Unsetenv("SERVER_LISTEN_PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("APP_NAME")
	}()

	config, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, "5555", config.ServerPort)
	assert.Equal(t, "postgres://localhost/test", config.DatabaseURL)
	assert.Equal(t, "testapp", config.AppName)
}

func TestConfigSingleton(t *testing.T) {
	ResetForTesting()
	setTestDefaults(t)

	os.Setenv("SERVER_LISTEN_PORT", "1234")
	defer os.Unsetenv("SERVER_LISTEN_PORT")

	config1, err := GetConfig()
	require.NoError(t, err)

	config2, err := GetConfig()
	require.NoError(t, err)

	config3 := MustGetConfig()

	assert.Equal(t, config1, config2)
	assert.Equal(t, config1, config3)
}
