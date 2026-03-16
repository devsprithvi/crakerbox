package main

import (
	"embed"
	"crackerboxd/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cmd.Assets = assets
	cmd.Execute()
}
