package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {

	var config Config

	config.MaxLogSize = 100
	config.ServiceName = "test"
	config.Level = "warn"
	config.NotDisplayLine = true
	Init(&config)
	Info("aaa")
	Error("bbb")
}
