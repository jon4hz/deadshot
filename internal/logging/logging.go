package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	Log    *logrus.Logger
	logDir string
)

func init() {
	folder, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	logDir = filepath.Join(folder, "deadshot", "logs")
	ensureFolderExists(logDir)

	Log = logrus.New()
	Log.Formatter = &logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "",
	}
	Log.Level = logrus.InfoLevel
	Log.SetReportCaller(true)
}

func SetFile() error {
	f := filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().UTC().Format("20060102150405")))

	logFile, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	Log.Out = logFile
	return nil
}

func ensureFolderExists(folder string) {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			panic(err)
		}
	}
}

func RemoveLogs() error {
	// remove all lof files from the logDir
	files, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(logDir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLastLog returns the content of the last log file.
func GetLastLog() (string, error) {
	file, err := os.ReadFile(getLastLogFile())
	if err != nil {
		return "", err
	}
	return string(file), nil
}

// getLastLogFile returns the last log file.
func getLastLogFile() string {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return ""
	}

	if len(files) == 0 {
		return ""
	}

	return filepath.Join(logDir, files[len(files)-1].Name())
}

// RemoveLastLog removes the last log file.
func RemoveLastLog() error {
	return os.Remove(getLastLogFile())
}
