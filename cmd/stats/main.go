package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/remind101/dockerstats"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "adapter",
		Value:  "log",
		EnvVar: "STAT_ADAPTER",
	},
	cli.StringFlag{
		Name:   "statsd.address",
		Value:  "localhost:8125",
		EnvVar: "STATSD_ADDR",
	},
	cli.StringFlag{
		Name:   "template",
		Value:  stats.L2MetTemplate,
		EnvVar: "STAT_TEMPLATE",
	},
	cli.IntFlag{
		Name:   "resolution",
		Value:  stats.DefaultResolution,
		EnvVar: "RESOLUTION",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "dockerstats"
	app.Flags = flags
	app.Action = run

	app.Run(os.Args)
}

func run(c *cli.Context) {
	stat, err := stats.New()
	must(err)

	stat.Adapter = newAdapter(c)
	stat.Resolution = c.Int("resolution")

	err = stat.Run()
	must(err)
}

func newAdapter(c *cli.Context) stats.Adapter {
	var (
		a   stats.Adapter
		err error
	)

	switch c.String("adapter") {
	case "log":
		a, err = stats.NewLogAdapter(c.String("template"))
	case "statsd":
		a, err = stats.NewStatsdAdapter(c.String("statsd.address"), c.String("template"))
	default:
		err = fmt.Errorf("unable to find an adapter matching: %s", c.String("adapter"))
	}

	must(err)
	return a
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
