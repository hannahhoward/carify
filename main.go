package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "carify",
		Usage:     "generate a car file from a regular file",
		ArgsUsage: "[inputFile outputCarFile]",
		Action: func(c *cli.Context) error {
			if c.Args().Len() != 2 {
				return errors.New("Wrong number of arguments")
			}
			inputFile, err := filepath.Abs(c.Args().Get(0))
			if err != nil {
				return err
			}
			outputFile, err := filepath.Abs(c.Args().Get(1))
			if err != nil {
				return err
			}
			return CarGenerator{
				InputFile:  inputFile,
				OutputFile: outputFile,
			}.Generate()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
