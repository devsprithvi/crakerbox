package main

import (
	"embed"
	"matchboxd/cmd"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cmd.Assets = assets
	cmd.Execute()
}
