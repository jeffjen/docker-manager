package main

import (
	dcli "github.com/jeffjen/go-discovery/cli"

	cli "github.com/codegangsta/cli"
)

var (
	Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "addr",
			Usage: "API endpoint for admin",
		},
		cli.StringFlag{
			Name:  "provider",
			Usage: "Provisioning provider",
			Value: "AWS",
		},
	}
)

func NewFlag() []cli.Flag {
	return append(Flags, dcli.Flags...)
}
