package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cdnjs/tools/npm"
	"github.com/cdnjs/tools/packages"
	"github.com/cdnjs/tools/util"
)

var (
	// Store the number of validation errors
	validationErrorCount uint = 0
)

func main() {
	flag.Parse()
	subcommand := flag.Arg(0)

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	if subcommand == "lint" {
		names := flag.Args()[1:]
		runAllChecks := len(names) == 1

		for _, name := range names {
			lintPackage(runAllChecks, name)
		}

		if validationErrorCount > 0 {
			fmt.Printf("%d linting error(s)\n", validationErrorCount)
			os.Exit(1)
		}
		return
	}

	panic("unknown subcommand")
}

func lintPackage(runAllChecks bool, path string) {
	ctx := util.ContextWithName(path)

	util.Debugf(ctx, "Linting %s...\n", path)

	pckg, readerr := packages.ReadPackageJSON(ctx, path)
	if readerr != nil {
		err(ctx, readerr.Error())
		return
	}

	if pckg.Name == "" {
		err(ctx, shouldNotBeEmpty(".name"))
	}

	if pckg.Version == "" {
		err(ctx, shouldBeEmpty(".version"))
	}

	// if pckg.NpmName != nil && *pckg.NpmName == "" {
	// 	err(ctx, shouldBeEmpty(".NpmName"))
	// }

	// if len(pckg.NpmFileMap) > 0 {
	// 	err(ctx, shouldBeEmpty(".NpmFileMap"))
	// }

	if pckg.Autoupdate != nil {
		if pckg.Autoupdate.Source != "npm" && pckg.Autoupdate.Source != "git" {
			err(ctx, "Unsupported .autoupdate.source: "+pckg.Autoupdate.Source)
		}
	} else {
		warn(ctx, ".autoupdate should not be null. Package will never auto-update")
	}

	if pckg.Repository.Repotype != "git" {
		err(ctx, "Unsupported .repository.type: "+pckg.Repository.Repotype)
	}

	if pckg.Autoupdate != nil && pckg.Autoupdate.Source == "npm" {
		if !npm.Exists(pckg.Autoupdate.Target) {
			err(ctx, "package doesn't exists on npm")
		} else {
			if runAllChecks {
				counts := npm.GetMonthlyDownload(pckg.Autoupdate.Target)
				if counts.Downloads < 800 {
					err(ctx, "package download per month on npm is under 800")
				}
			}
		}
	}

}

func err(ctx context.Context, s string) {
	util.Printf(ctx, "error: "+s)
	validationErrorCount += 1
}

func warn(ctx context.Context, s string) {
	util.Printf(ctx, "warning: "+s)
}

func shouldBeEmpty(name string) string {
	return fmt.Sprintf("%s should be empty\n", name)
}

func shouldNotBeEmpty(name string) string {
	return fmt.Sprintf("%s should be specified\n", name)
}