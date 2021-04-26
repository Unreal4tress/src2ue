package cmd

import (
	"os"
	"path/filepath"

	"github.com/Unreal4tress/go-sourceformat/vmf"
	ulc "github.com/Unreal4tress/uelevelclip"
	"github.com/fatih/color"

	"github.com/spf13/cobra"
)

const gScale = 2.0

var rootCmd = &cobra.Command{
	Use:   "src2ue [VMF]",
	Short: "Transform vmf to UE clipboard data",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		inFile := args[0]
		if filepath.Ext(inFile) != ".vmf" {
			color.Red("Input file is not *.vmf file")
			return
		}
		inF, err := os.Open(inFile)
		if err != nil {
			color.Red("Cannot open the file")
			return
		}
		defer inF.Close()
		vmf, err := vmf.NewDecoder(inF).Decode()
		if err != nil {
			color.Red("Failed to decode vmf file")
			return
		}
		outFile := inFile[0:len(inFile)-3] + "uecb.txt"
		outF, err := os.Create(outFile)
		if err != nil {
			color.Red("Failed to create new file")
			return
		}
		defer outF.Close()
		data := transform(vmf)
		if data == nil {
			return
		}
		if err := ulc.NewEncoder(outF, nil).Encode(data); err != nil {
			color.Red("Failed to output data")
		}
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
