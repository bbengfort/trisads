package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/bbengfort/trisads"
	"github.com/bbengfort/trisads/pb"
	"github.com/bbengfort/trisads/store"
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
			Usage:  "run the trisa directory service",
			Action: serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a, addr",
					Usage: "the address and port to bind the server on",
					Value: ":4433",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "path to LevelDB directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
			},
		},
		{
			Name:      "load",
			Usage:     "load the directory from a csv file",
			Action:    load,
			ArgsUsage: "csv [csv ...]",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "dsn to connect to trisa directory storage",
					EnvVar: "TRISADS_DATABASE",
				},
			},
		},
	}

	app.Run(os.Args)
}

func serve(c *cli.Context) (err error) {
	return nil
}

// Quick helper function to load the LevelDB database with initial directory info.
func load(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return cli.NewExitError("specify path to csv data to load", 1)
	}

	var dbpath string
	if dbpath = c.String("db"); dbpath == "" {
		return cli.NewExitError("please specify path to LevelDB storage", 1)
	}

	var db store.Store
	if db, err = store.Open(dbpath); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer db.Close()

	for _, path := range c.Args() {
		var f *os.File
		if f, err = os.Open(path); err != nil {
			return cli.NewExitError(err, 1)
		}

		rows := 0
		reader := csv.NewReader(f)

		for {
			var record []string
			if record, err = reader.Read(); err != nil {
				break
			}

			rows++
			if rows == 1 {
				// Skip the expected header: entity name,country,category,url,address
				// TODO: validate the header
				continue
			}

			vasp := pb.VASP{
				VaspEntity: &pb.Entity{
					VaspFullLegalName:    record[0],
					VaspFullLegalAddress: record[4],
					VaspURL:              record[3],
					VaspCategory:         record[2],
					VaspCountry:          record[1],
				},
			}

			var id uint64
			if id, err = db.Create(vasp); err != nil {
				return cli.NewExitError(err, 1)
			}

			var check pb.VASP
			if check, err = db.Retrieve(id); err != nil {
				return cli.NewExitError(err, 1)
			}

			data, _ := json.Marshal(check)
			fmt.Println(string(data))

		}

		f.Close()
		if err != nil && err != io.EOF {
			return cli.NewExitError(err, 1)
		}
	}

	return nil
}
