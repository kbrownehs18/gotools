package tests

import (
	"testing"

	"github.com/kbrownehs18/gotools/log"
)

func TestLog(t *testing.T) {
	logger, err := log.NewLogger("Test", "console", "error")
	if err != nil {
		t.Error(err)
	}
	logger.Info("Test info message info")

	logger.Error("Test info message error")
}

func TestLogFile(t *testing.T) {
	logger, err := log.NewLogger("Test", "file", "error")
	if err != nil {
		t.Error(err)
	}
	logger.Info("Test info message info")

	logger.Error("Test info message error")
}

func BenchmarkLog(b *testing.B) {
	b.ResetTimer()
	num := 20 << 5
	fileHandler, err := log.NewFileHandler("./logs", "error.log", "daily", 20)
	if err != nil {
		b.Error(err)
	}
	logger, err := log.NewLogger("Test", "file", "error", fileHandler)
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < num; i++ {
		logger.Error("Test info message error benchmark")
	}
}
