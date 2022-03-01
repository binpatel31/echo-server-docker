package main

import (
	"embed"
	"github.com/binpatel31/echo-server-docker/cmd/echo-server"
)

//go:embed templates
var wwwFiles embed.FS

func main() {
	echo_server.TemplateFiles = wwwFiles
	echo_server.Run()
}
