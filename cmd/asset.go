package cmd

import (
	"encoding/json"
	"os"

	cprint "github.com/fatih/color"
)

var assets struct {
	Materials map[string]struct {
		Asset string `json:"asset"`
		W     int    `json:"w"`
		H     int    `json:"h"`
	} `json:"materials"`
	Props map[string]struct {
		Asset string `json:"asset"`
	} `json:"props"`
}

var unknownMaterials map[string]struct{}

func loadAssets() {
	f, err := os.Open("assets.json")
	if err != nil {
		cprint.Yellow("Warning: No found \"assets.json\"")
		return
	}
	if err := json.NewDecoder(f).Decode(&assets); err != nil {
		cprint.Yellow("Warning: Failed to parse \"assets.json\"")
		return
	}
}
