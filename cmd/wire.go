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
	Use:   "wire [target] [directory]",
	Short: "Generate wiring code",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {
		config, err := helpers.ReadConfig()
		if err != nil {
			fatalDixError(err)
		}

		targetDir := "."

		if len(args) > 1 {
			targetDir = args[0]
		}

		p := parser.NewParser()
		g := generator.NewGenerator()
		mt, err := p.Parse(targetDir)
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

func init() {
	rootCmd.AddCommand(wireCmd)

}
