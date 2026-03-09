package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/smtdfc/dix/generator"
	"github.com/smtdfc/dix/helpers"
	"github.com/smtdfc/dix/parser"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [directory]",
	Short: "Scan source code and generate dependency injection wiring",
	Long: `The 'run' command performs a full analysis of your Go source code 
within the specified directory. 

Example:
  dix run ./internal/app`,

	Run: func(cmd *cobra.Command, args []string) {
		targetDir := "."
		if len(args) > 0 {
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

		err = helpers.WriteTextFile(code, "./generated/dix/root.go")
		if err != nil {
			fatalDixError(err)
		}

		fmt.Printf("\033[32m[Run]\033[0m Running ... \n ")
		command := exec.Command("go", "run", ".")

		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fatalDixError(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)

}
