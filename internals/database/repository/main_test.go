package repository

import (
	"os"
	"testing"

	"catgoose/harmony/internals/logger"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}
