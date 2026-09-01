// Copyright (C) 2025 T-Force I/O
// This file is part of TFprism
//
// TFprism is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFprism is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFprism. If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/tforceaio/tf-prism/tui"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logFileName      = "prism.log"
	logFileMaxSize   = 10 // megabytes per file before rotation.
	logFileMaxBackup = 3  // maximum rotated files to retain.
	logFileMaxAge    = 28 // maximum days to retain old log files.
)

// Entrypoint for creating a ZeroLog logger instance.
func InitZerolog(configDir string, useFS bool) (zerolog.Logger, *lumberjack.Logger, error) {
	level := resolveLevel()
	zerolog.SetGlobalLevel(level)

	colorSupported := tui.IsTTY()
	consoleWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: zerolog.ConsoleWriter{Out: os.Stdout, NoColor: !colorSupported, TimeFormat: time.DateTime},
		},
		Level: zerolog.TraceLevel,
	}

	logFile, err := InitLogFile(useFS, configDir)
	if logFile == nil {
		consoleLogger := zerolog.New(consoleWriter).With().Timestamp().Logger()
		return consoleLogger, nil, err
	}

	fileWriter := &zerolog.FilteredLevelWriter{
		Writer: zerolog.LevelWriterAdapter{
			Writer: logFile,
		},
		Level: zerolog.TraceLevel,
	}
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	return logger, logFile, nil
}

// Create and return rotating log file writer only if useFS is true.
func InitLogFile(useFS bool, workdingDir string) (*lumberjack.Logger, error) {
	if !useFS {
		return nil, nil
	}
	logDir := workdingDir
	if logDir == "" {
		logDir = "."
	}
	logDir = filepath.Join(logDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	return &lumberjack.Logger{
		Filename:   filepath.Join(logDir, logFileName),
		MaxSize:    logFileMaxSize,
		MaxBackups: logFileMaxBackup,
		MaxAge:     logFileMaxAge,
	}, nil
}

// Parse configured log level, fallback to info when invalid or missing.
func resolveLevel() zerolog.Level {
	if cfg == nil || cfg.Logger == nil {
		return zerolog.InfoLevel
	}
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return level
}
