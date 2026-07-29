package main

import (
	"flag"
	"log"

	"github.com/local/wifi-share/internal/app"
)

func main() {
	config, err := loadLocalConfig("config.local.json")
	if err != nil {
		log.Fatal(err)
	}
	address := flag.String("addr", config.Address, "HTTP listen address")
	root := flag.String("root", config.Root, "directory exposed to connected devices")
	data := flag.String("data", config.Data, "application data directory")
	web := flag.String("web", config.Web, "compiled web application directory")
	flag.Parse()

	server, err := app.New(app.Config{
		Address:  *address,
		ShareDir: *root,
		DataDir:  *data,
		WebDir:   *web,
		Password: config.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("WiFi Share is ready. Open one of these URLs:")
	for _, address := range accessURLs(*address) {
		log.Printf("  %s", address)
	}
	runApplication(server, config, accessURLs(*address))
}
