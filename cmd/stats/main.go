package main

import (
	"log"
	"os"

	stats "github.com/remind101/dockerstats"
)

func main() {
	stat, err := stats.New(os.Getenv("DOCKER_HOST"))
	if err != nil {
		log.Fatal(err)
	}

	a, err := stats.NewLogAdapter(os.Getenv("STAT_TEMPLATE"))
	if err != nil {
		log.Fatal(err)
	}

	stat.Adapter = a

	if err := stat.Run(); err != nil {
		log.Fatal(err)
	}
}
