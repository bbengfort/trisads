package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/bbengfort/trisads"
	"github.com/bbengfort/trisads/pb"
	"github.com/bbengfort/trisads/store"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	client pb.TRISADirectoryClient
)

func main() {
	app := cli.NewApp()

	app.Name = "trisads"
	app.Version = trisads.Version()
	app.Usage = "a gRPC based directory service for TRISA identity lookups"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "e, endpoint",
			Usage:  "the url to connect the directory service client",
			Value:  "vaspdirectory.net:443",
			EnvVar: "TRISA_DIRECTORY_URL",
		},
		cli.BoolFlag{
			Name:  "S, no-secure",
			Usage: "do not connect via TLS (e.g. for development)",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:     "serve",
			Usage:    "run the trisa directory service",
			Category: "server",
			Action:   serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "a, addr",
					Usage:  "the address and port to bind the server on",
					EnvVar: "TRISADS_BIND_ADDR",
				},
			},
		},
		{
			Name:      "load",
			Usage:     "load the directory from a csv file",
			Category:  "server",
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
		{
			Name:     "verify",
			Usage:    "mark a VASP entity as verified and create certificates",
			Category: "admin",
			Action:   verify,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l, list",
					Usage: "list VASPs that require verification and exit",
				},
				cli.Uint64Flag{
					Name:  "v, vasp",
					Usage: "the ID of the VASP to mark as verified",
				},
			},
		},
		{
			Name:     "register",
			Usage:    "register a VASP using json data",
			Category: "client",
			Action:   register,
			Before:   initClient,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d, data",
					Usage: "the json file containing the VASP data record",
				},
				cli.BoolFlag{
					Name:  "V, no-verify",
					Usage: "mark the request as no verification required",
				},
			},
		},
		{
			Name:     "lookup",
			Usage:    "lookup VASPs using name or ID",
			Category: "client",
			Action:   lookup,
			Before:   initClient,
			Flags: []cli.Flag{

				cli.StringFlag{
					Name:  "n, name",
					Usage: "name of the VASP to lookup (case-insensitive, exact match)",
				},
				cli.Uint64Flag{
					Name:  "i, id",
					Usage: "id of the VASP to lookup",
				},
			},
		},
		{
			Name:     "search",
			Usage:    "search for VASPs using name or country",
			Category: "client",
			Action:   search,
			Before:   initClient,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "n, name",
					Usage: "one or more names of the VASPs to search for",
				},
				cli.StringSliceFlag{
					Name:  "c, country",
					Usage: "one or more countries of the VASPs to search for",
				},
			},
		},
	}

	app.Run(os.Args)
}

// Serve the TRISA directory service
func serve(c *cli.Context) (err error) {
	var conf *trisads.Settings
	if conf, err = trisads.Config(); err != nil {
		return cli.NewExitError(err, 1)
	}

	if addr := c.String("addr"); addr != "" {
		conf.BindAddr = addr
	}

	var srv *trisads.Server
	if srv, err = trisads.New(conf); err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = srv.Serve(); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// Load the LevelDB database with initial directory info from CSV
// TODO: remove or make more robust
func load(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return cli.NewExitError("specify path to csv data to load", 1)
	}

	var dsn string
	if dsn = c.String("db"); dsn == "" {
		return cli.NewExitError("please specify a dsn to connect to the directory store", 1)
	}

	var db store.Store
	if db, err = store.Open(dsn); err != nil {
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

// Verify a registered entity and assign keys (server-side CLI)
func verify(c *cli.Context) (err error) {
	return cli.NewExitError("not implemented", 7)
}

// Register an entity using the API from a CLI client
func register(c *cli.Context) (err error) {
	req := &pb.RegisterRequest{
		Verify: !c.Bool("no-verify"),
	}

	var path string
	if path = c.String("data"); path == "" {
		return cli.NewExitError("specify a json file to load the entity data from", 1)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	if err = json.Unmarshal(data, &req.Entity); err != nil {
		return cli.NewExitError(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rep, err := client.Register(ctx, req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return printJSON(rep)
}

// Lookup VASPs using the API from a CLI client
func lookup(c *cli.Context) (err error) {
	name := c.String("name")
	id := c.Uint64("id")

	if name == "" && id == 0 {
		return cli.NewExitError("specify either name or id for lookup", 1)
	}

	if name != "" && id > 0 {
		return cli.NewExitError("specify either name or id for lookup, not both", 1)
	}

	req := &pb.LookupRequest{
		Name: name,
		Id:   id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rep, err := client.Lookup(ctx, req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return printJSON(rep)
}

// Search for VASPs by name or country using the API from a CLI client
func search(c *cli.Context) (err error) {
	req := &pb.SearchRequest{
		Name:    c.StringSlice("name"),
		Country: c.StringSlice("country"),
	}

	if len(req.Name) == 0 && len(req.Country) == 0 {
		return cli.NewExitError("specify search query", 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rep, err := client.Search(ctx, req)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	return printJSON(rep)
}

// helper function to create the GRPC client with default options
func initClient(c *cli.Context) (err error) {
	var opts []grpc.DialOption
	if c.GlobalBool("no-secure") {
		opts = append(opts, grpc.WithInsecure())
	} else {
		config := &tls.Config{}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	}

	var cc *grpc.ClientConn
	if cc, err = grpc.Dial(c.GlobalString("endpoint"), opts...); err != nil {
		return cli.NewExitError(err, 1)
	}
	client = pb.NewTRISADirectoryClient(cc)
	return nil
}

// helper function to print JSON response and exit
func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	fmt.Println(string(data))
	return nil
}
