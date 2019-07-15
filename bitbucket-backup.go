package main

import (
	"os"

	flags "github.com/jessevdk/go-flags"
)

type Options struct {
	Username string `short:"u" long:"username" description:"bitbucket username"`
	Password string `short:"p" long:"password" description:"bitbucket user's password"`
	Location string `short:"l" long:"location" description:"local backup location"`
}

func main() {
	var opts Options

	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
