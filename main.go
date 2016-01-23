package main

import (
	disc "github.com/jeffjen/go-discovery"
	dcli "github.com/jeffjen/go-discovery/cli"
	web "github.com/jeffjen/podd/web"

	log "github.com/Sirupsen/logrus"
	cli "github.com/codegangsta/cli"

	"os"
)

const (
	ManagerPrefix = "/master/nodes"
)

func main() {
	app := cli.NewApp()
	app.Name = "podd"
	app.Usage = "Facilitate management of service provision by docker swarm"
	app.Authors = []cli.Author{
		cli.Author{"Yi-Hung Jen", "yihungjen@gmail.com"},
	}
	app.Flags = NewFlag()
	app.Action = Manager
	app.Run(os.Args)
}

func Manager(ctx *cli.Context) {
	var (
		addr = ctx.String("addr")
	)

	// setup register path for discovery
	disc.RegisterPath = ManagerPrefix

	if err := dcli.Before(ctx); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("halt")
	}

	if addr == "" {
		log.WithFields(log.Fields{"err": "Required flag addr missing"}).Fatal("halt")
	}

	log.WithFields(log.Fields{"addr": addr}).Info("API endpoint begin")
	web.RunAPIEndpoint(addr)
}
