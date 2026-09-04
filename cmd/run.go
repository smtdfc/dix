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
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := helpers.ReadConfig()
		if err != nil {
			fatalDixError(err)
		}

		targetDir := "."
		var goRunArgs []string

		if len(args) > 0 {
			targetDir = args[0]
			if len(args) > 1 {
				goRunArgs = args[1:]
			}
		}

		p := parser.NewParser()
		g := generator.NewGenerator()
		mt, err := p.ParseWithOptions(targetDir, parser.ScanOptions{Workspace: runWorkspace, NoCache: runNoCache})
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

		fmt.Printf("\033[32m[Run]\033[0m Running ... \n")

		cmdArgs := []string{"run", "."}
		cmdArgs = append(cmdArgs, goRunArgs...)

		command := exec.Command("go", cmdArgs...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fatalDixError(err)
		}
	},
}

var runWorkspace bool
var runNoCache bool

func init() {
	runCmd.Flags().BoolVar(&runWorkspace, "workspace", false, "scan every module declared in the nearest go.work file")
	runCmd.Flags().BoolVar(&runNoCache, "no-cache", false, "ignore and do not write the scan cache")
	rootCmd.AddCommand(runCmd)

}
