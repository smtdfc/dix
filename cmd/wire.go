package cmd

import (
	"fmt"
	"time"

	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/helpers"
	"github.com/smtdfc/dix/parser"
	"github.com/spf13/cobra"
)

var wireCmd = &cobra.Command{
	Use:   "wire [directory]",
	Short: "Generate wiring code",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {
		config, err := helpers.ReadConfig()
		if err != nil {
			fatalDixError(err)
		}

		targetDir := "."

		if len(args) > 0 {
			targetDir = args[0]
		}

		p := parser.NewParser()
		g := generator.NewGenerator()
		mt, err := p.ParseWithOptions(targetDir, parser.ScanOptions{Workspace: wireWorkspace, NoCache: wireNoCache})
		if err != nil {
			fatalDixError(err)
		}

		now := time.Now().Unix()
		fileName := fmt.Sprintf("scan_%d.dix", now)
		err = helpers.SaveMetadata(mt, fileName)
		if err != nil {
			fatalDixError(err)
		}

		code, err := g.Generate(mt)
		if err != nil {
			fatalDixError(err)
		}

		outputPath := "./generated/dix/root.go"
		if config.Output != "" {
			outputPath = config.Output
		}
		err = helpers.WriteTextFile(code, outputPath)
		if err != nil {
			fatalDixError(err)
		}

	},
}

var wireWorkspace bool
var wireNoCache bool

func init() {
	wireCmd.Flags().BoolVar(&wireWorkspace, "workspace", false, "scan every module declared in the nearest go.work file")
	wireCmd.Flags().BoolVar(&wireNoCache, "no-cache", false, "ignore and do not write the scan cache")
	rootCmd.AddCommand(wireCmd)

}
