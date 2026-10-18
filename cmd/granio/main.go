// Run a command line helper for various Fyne tools.
//
// Deprecated: Install github.com/Uzer007-admin/granio-tools/cmd/granio for latest version.
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Uzer007-admin/granio/v2/cmd/granio/internal/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	fmt.Println("Утилиты Granio развиваются также в отдельном репозитории granio-tools.")

	app := &cli.App{
		Name:        "granio",
		Usage:       "Сборка и упаковка приложений Granio.",
		Description: "Утилита разработки, сборки и распространения приложений Granio.",
		Commands: []*cli.Command{
			commands.Bundle(),
			commands.Env(),
			commands.Get(),
			commands.Install(),
			commands.Package(),
			commands.Release(),
			commands.Version(),
			commands.Serve(),
			commands.Translate(),
			commands.Build(),

			// Deprecated: Use "go mod vendor" instead.
			commands.Vendor(), //lint:ignore SA1019 This whole tool is deprecated.
		},
	}

	if info, ok := debug.ReadBuildInfo(); !ok {
		app.Version = "could not retrieve version information (ensure module support is activated and build again)"
	} else {
		app.Version = info.Main.Version
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
