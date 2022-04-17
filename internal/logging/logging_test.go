package logging

import (
	"testing"
	"time"
)

func TestGetTime(t *testing.T) {
	current_time := time.Now().UTC().Format("20060102150405")
	t.Log(current_time)
}

func TestGetLastLog(t *testing.T) {
	err := SetFile()
	if err != nil {
		t.Error(err)
	}
	Log.Info("testing")
	_, err = GetLastLog()
	if err != nil {
		t.Error(err)
	}
}

func TestRemoveLogs(t *testing.T) {
	err := SetFile()
	if err != nil {
		t.Error(err)
	}
	Log.Info("testing")
	err = RemoveLogs()
	if err != nil {
		t.Error(err)
	}
}
