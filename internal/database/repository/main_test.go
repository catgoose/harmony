package repository

import (
	"os"
	"testing"

	"catgoose/harmony/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init()
	os.Exit(m.Run())
}
