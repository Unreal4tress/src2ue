package cmd

import (
	"os"
	"path/filepath"

	"github.com/Unreal4tress/go-sourceformat/vmf"
	ulc "github.com/Unreal4tress/uelevelclip"
	cprint "github.com/fatih/color"
	"github.com/spf13/cobra"
)

const gScale = 2.0

var rootCmd = &cobra.Command{
	Use:   "src2ue [RULE]",
	Short: "Transform vmf to UE clipboard data",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if loadRule(args[0]) != nil {
			return
		}
		if filepath.Ext(rule.SourceFile) != ".vmf" {
			cprint.Red("Input file is not *.vmf file")
			return
		}
		inF, err := os.Open(rule.SourceFile)
		if err != nil {
			cprint.Red("Cannot open the file")
			return
		}
		defer inF.Close()
		vmf, err := vmf.NewDecoder(inF).Decode()
		if err != nil {
			cprint.Red("Failed to decode vmf file")
			return
		}
		outFile := rule.DstPath + rule.MapName + ".uecb.txt"
		outF, err := os.Create(outFile)
		if err != nil {
			cprint.Red("Failed to create new file")
			return
		}
		defer outF.Close()
		loadAssets()
		data := transform(vmf)
		if data == nil {
			return
		}
		if err := ulc.NewEncoder(outF, nil).Encode(data); err != nil {
			cprint.Red("Failed to output data")
		}
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
