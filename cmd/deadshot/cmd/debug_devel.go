//go:build devel
// +build devel

package cmd

import (
	gops "github.com/google/gops/agent"
	gosivy "github.com/nakabonne/gosivy/agent"
)

func startDebugAgents() error {
	var err error
	block := make(chan struct{})
	go func() {
		err = gosivy.Listen(gosivy.Options{})
		if err != nil {
			close(block)
			return
		}
		defer gosivy.Close()
		<-block
	}()
	go func() {
		err = gops.Listen(gops.Options{})
		if err != nil {
			close(block)
			return
		}
		defer gops.Close()
		<-block
	}()
	return nil
}
