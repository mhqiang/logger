package logger

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {

	InitDefaultLogger()
	config := DefaultConfig

	fmt.Println(config.LogPath, config.Compress, config.NotDisplayLine, config.Stdout)
	Info("aaa")
	Error("bbb")

}
