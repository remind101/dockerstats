package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"github.com/quipo/statsd"
	"github.com/remind101/dockerstats"
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:   "url",
		Value:  "log://",
		EnvVar: "STAT_URL",
	},
	cli.StringFlag{
		Name:   "template",
		Value:  stats.L2MetTemplate,
		EnvVar: "STAT_TEMPLATE",
	},
	cli.StringSliceFlag{
		Name:   "whitelist",
		Value:  &cli.StringSlice{},
		EnvVar: "STAT_WHITELIST",
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
	stat.Whitelist = c.StringSlice("whitelist")

	err = stat.Run()
	must(err)
}

func newAdapter(c *cli.Context) stats.Adapter {
	var (
		a   stats.Adapter
		err error
	)

	u, err := url.Parse(c.String("url"))
	must(err)

	switch u.Scheme {
	case "log":
		a, err = stats.NewLogAdapter(c.String("template"), nil)
	case "statsd":
		client := statsd.NewStatsdClient(u.Host, "")
		err = client.CreateSocket()
		must(err)
		a, err = stats.NewStatsdAdapter(client, c.String("template"))
	default:
		err = fmt.Errorf("unable to find an adapter to handle: %s", c.String("url"))
	}

	must(err)
	return a
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
