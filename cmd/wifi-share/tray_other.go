//go:build !windows

package main

import (
	"log"

	"github.com/local/wifi-share/internal/app"
)

func runApplication(server *app.App, _ localConfig, _ []string) {
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
