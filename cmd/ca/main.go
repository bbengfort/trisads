package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/bbengfort/trisads"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "ca"
	app.Version = trisads.Version()
	app.Usage = "a pseudo certificate authority for testing purposes"
	app.Flags = []cli.Flag{}
	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "create CA certs and keys if they do not exist",
			Action: initCA,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "c, certs",
					Usage:  "local directory where certificates and keys are stored",
					Value:  "fixtures/certs",
					EnvVar: "TRISA_CA_CERTS",
				},
				cli.BoolFlag{
					Name:  "f, force",
					Usage: "overwrite keys even if they already exist",
				},
			},
		},
	}

	app.Run(os.Args)
}

func initCA(c *cli.Context) (err error) {
	force := c.Bool("force")
	certPath := filepath.Join(c.String("certs"), "ca.crt")
	keyPath := filepath.Join(c.String("certs"), "ca.key")

	if !force {
		if _, err = os.Stat(certPath); err == nil {
			return cli.NewExitError("certificate file already exists", 1)
		}
		if _, err = os.Stat(keyPath); err == nil {
			return cli.NewExitError("private key file already exists", 1)
		}
	}

	// Create a certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1942),
		Subject: pkix.Name{
			Organization:  []string{"TRISA Test Net"},
			Country:       []string{"US"},
			Province:      []string{"MD"},
			Locality:      []string{"Queenstown"},
			StreetAddress: []string{"215 Alynn Way"},
			PostalCode:    []string{"21658"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create private key
	priv, _ := rsa.GenerateKey(rand.Reader, 4096)
	pub := &priv.PublicKey
	signed, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("create ca failed: %s", err), 1)
	}

	// Save the key to a file
	var cf, kf *os.File
	if cf, err = os.Create(certPath); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer cf.Close()
	if err = pem.Encode(cf, &pem.Block{Type: "CERTIFICATE", Bytes: signed}); err != nil {
		return cli.NewExitError(err, 1)
	}

	if kf, err = os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return cli.NewExitError(err, 1)
	}
	defer kf.Close()
	if err = pem.Encode(kf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return cli.NewExitError(err, 1)
	}

	return nil
}
