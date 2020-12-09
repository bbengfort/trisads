package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bbengfort/trisads"
	"github.com/bbengfort/trisads/sectigo"
	"github.com/urfave/cli"
)

var (
	api     *sectigo.Sectigo
	encoder *json.Encoder
)

func main() {
	app := cli.NewApp()

	app.Name = "sectigo"
	app.Version = trisads.Version()
	app.Usage = "CLI helper for Sectigo API access and debugging"
	app.Before = initAPI
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "u, username",
			Usage:  "API access login username",
			EnvVar: sectigo.UsernameEnv,
		},
		cli.StringFlag{
			Name:   "p, password",
			Usage:  "API access login password",
			EnvVar: sectigo.PasswordEnv,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:   "auth",
			Usage:  "check authentication status with server",
			Action: auth,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "d, debug",
					Usage: "do not refresh or authenticate, print state and exit",
				},
				cli.BoolFlag{
					Name:  "C, cache",
					Usage: "print cache location and exit",
				},
			},
		},
		{
			Name:   "licenses",
			Usage:  "view the ordered/issued certificates",
			Action: licenses,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "authorities",
			Usage:  "view the current users authorities by ecosystem",
			Action: authorities,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "profiles",
			Usage:  "view profiles available to the user",
			Action: profiles,
			Flags:  []cli.Flag{},
		},
	}

	app.Run(os.Args)
}

func initAPI(c *cli.Context) (err error) {
	if api, err = sectigo.New(c.String("username"), c.String("password")); err != nil {
		return cli.NewExitError(err, 1)
	}

	encoder = json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	return nil
}

func auth(c *cli.Context) (err error) {
	creds := api.Creds()

	if c.Bool("cache") {
		if cacheFile := creds.CacheFile(); cacheFile != "" {
			fmt.Println(cacheFile)
		} else {
			fmt.Println("no credentials cache file exists")
		}
		return nil
	}

	if c.Bool("debug") {
		if creds.Valid() {
			fmt.Printf("credentials are valid until %s\n", creds.ExpiresAt)
			return nil
		}

		if creds.Current() {
			fmt.Printf("credentials are current until %s\n", creds.RefreshBy)
			return nil
		}

		fmt.Println("credentials are expired or invalid")
		return nil
	}

	if !creds.Valid() {
		if creds.Refreshable() {
			if err = api.Refresh(); err != nil {
				return cli.NewExitError(err, 1)
			}
		} else {
			if err = api.Authenticate(); err != nil {
				return cli.NewExitError(err, 1)
			}
		}
	}

	fmt.Println("user authenticated and credentials cached")
	return nil
}

func licenses(c *cli.Context) (err error) {
	var rep *sectigo.LicensesUsedResponse
	if rep, err = api.LicensesUsed(); err != nil {
		return cli.NewExitError(err, 1)
	}

	printJSON(rep)
	return nil
}

func authorities(c *cli.Context) (err error) {
	var rep []*sectigo.AuthorityResponse
	if rep, err = api.UserAuthorities(); err != nil {
		return cli.NewExitError(err, 1)
	}

	printJSON(rep)
	return nil
}

func profiles(c *cli.Context) (err error) {
	var rep []*sectigo.ProfileResponse
	if rep, err = api.Profiles(); err != nil {
		return cli.NewExitError(err, 1)
	}

	printJSON(rep)
	return nil
}

func printJSON(data interface{}) (err error) {
	if err = encoder.Encode(data); err != nil {
		return err
	}
	return nil
}
