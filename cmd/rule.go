package cmd

import (
	"encoding/json"
	"errors"
	"os"

	cprint "github.com/fatih/color"
)

var rule struct {
	MapName    string `json:"map_name"`
	GamePath   string `json:"game_path"`
	SourceFile string `json:"source_file"`
	AssetsFile string `json:"assets_file"`
	DstPath    string `json:"dst_path"`
}

func loadRule(path string) error {
	f, err := os.Open(path)
	if err != nil {
		cprint.Red("Error: No found rule file.")
		return errors.New("")
	}
	if err := json.NewDecoder(f).Decode(&rule); err != nil {
		cprint.Yellow("Error: Failed to parse rule file.")
		return errors.New("")
	}
	os.MkdirAll(rule.DstPath+rule.MapName+"/Brushes", os.ModePerm)
	return nil
}
