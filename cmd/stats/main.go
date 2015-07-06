package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/remind101/dockerstats"
)

var flags = []cli.Flag{
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

	a, err := stats.NewLogAdapter(c.String("template"))
	must(err)

	stat.Adapter = a
	stat.Resolution = c.Int("resolution")

	err = stat.Run()
	must(err)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
