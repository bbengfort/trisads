package main

import (
	"os"

	"github.com/bbengfort/trisads"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "trisads"
	app.Version = trisads.Version()
	app.Usage = "a gRPC based directory service for TRISA identity lookups"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:   "serve",
			Usage:  "run the olsen report server to view the data",
			Action: serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "the address and port to bind the server on",
					Value: ":3000",
				},
			},
		},
	}

	app.Run(os.Args)
}

func serve(c *cli.Context) (err error) {
	return nil
}
