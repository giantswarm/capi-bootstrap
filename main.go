package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/capi-bootstrap/cmd"
)

func main() {
	err := mainE()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", microerror.Pretty(err, true))
		os.Exit(2)
	}
}

func mainE() error {
	var logger micrologger.Logger
	{
		var err error
		config := micrologger.Config{}
		logger, err = micrologger.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var rootCommand *cobra.Command
	{
		var err error
		config := cmd.Config{
			Logger: logger,
		}
		rootCommand, err = cmd.New(config)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err := rootCommand.Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
