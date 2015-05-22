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

	stat.Adapter = stats.NewL2MetAdapter()

	if err := stat.Run(); err != nil {
		log.Fatal(err)
	}
}
