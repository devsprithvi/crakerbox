package cmd

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Assets holds the embedded frontend files, set from main.go
var Assets embed.FS

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the Matchbox dashboard UI",
	Long:  `Opens the Matchbox graphical dashboard powered by Wails.`,
	Run: func(cmd *cobra.Command, args []string) {
		app := NewDashboardApp()

		err := wails.Run(&options.App{
			Title:  "Matchbox Dashboard",
			Width:  1280,
			Height: 800,
			AssetServer: &assetserver.Options{
				Assets: Assets,
			},
			BackgroundColour: &options.RGBA{R: 18, G: 18, B: 24, A: 1},
			OnStartup:        app.startup,
			Bind: []interface{}{
				app,
			},
		})

		if err != nil {
			fmt.Println("Error launching UI:", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
