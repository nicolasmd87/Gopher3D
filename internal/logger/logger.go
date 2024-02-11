package logger

import (
	"embed"
	"encoding/json"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed zap-config.json
var configFile embed.FS

// Global logger accessible from other packages
var Log *zap.Logger

func Init() {
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", os.ModePerm) // Create the directory with appropriate permissions
	}

	configData, err := configFile.ReadFile("zap-config.json")
	if err != nil {
		panic("Failed to read embedded log configuration file: " + err.Error())
	}

	var cfg zap.Config
	if err := json.Unmarshal(configData, &cfg); err != nil {
		panic("Failed to unmarshal log configuration: " + err.Error())
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	Log = logger

	zap.ReplaceGlobals(Log)

	Log.Info("Logger initialized.")
}
