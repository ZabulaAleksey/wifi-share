package main

import (
	"flag"
	"log"

	"github.com/local/wifi-share/internal/app"
)

func main() {
	address := flag.String("addr", ":8080", "HTTP listen address")
	root := flag.String("root", "./shared", "directory exposed to connected devices")
	data := flag.String("data", "./data", "application data directory")
	web := flag.String("web", "./web/dist", "compiled web application directory")
	flag.Parse()

	server, err := app.New(app.Config{
		Address:  *address,
		ShareDir: *root,
		DataDir:  *data,
		WebDir:   *web,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Printf("WiFi Share is available at http://localhost%s", *address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
