package trisads

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/bbengfort/trisads/pb"
)

// Global issuer for TRISA CA (temporary).
var issuer = pkix.Name{
	CommonName:         "TRISA CA",
	Organization:       []string{"TRISA"},
	OrganizationalUnit: []string{"TRISA Test Net"},
	StreetAddress:      []string{},
	Locality:           []string{"Queenstown"},
	Province:           []string{"MD"},
	PostalCode:         []string{"21658"},
	Country:            []string{"United States"},
}

// CA key material for signing; lazily loaded from disk once.
var (
	catls tls.Certificate
	ca    *x509.Certificate
)

// LoadCACerts retrieves the CA certificate and private key to sign new certificates.
func LoadCACerts(certFile, keyFile string) (err error) {
	if catls, err = tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		return err
	}

	if ca, err = x509.ParseCertificate(catls.Certificate[0]); err != nil {
		return err
	}

	return nil
}

// Create a certificate for the specified entity with TRISA CA default setetings.
func createCertficate(name *pb.Name) (cert *x509.Certificate, err error) {
	if ca == nil {
		return nil, errors.New("certificate authority has not been loaded")
	}

	cert = &x509.Certificate{
		Subject: pkix.Name{
			CommonName:         name.CommonName,
			Organization:       []string{name.Organization},
			OrganizationalUnit: []string{name.OrganizationalUnit},
			StreetAddress:      []string{},
			Locality:           []string{name.Locality},
			Province:           []string{name.StateProvince, name.IncStateProvince},
			PostalCode:         []string{},
			Country:            []string{name.CountryRegion, name.IncCountryRegion},
		},
		Issuer:      issuer,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	if cert.SerialNumber, err = newSerialNumber(); err != nil {
		return nil, err
	}

	return cert, nil
}

// Create a random serial number for use with the certificate. Serial numbers are large
// enough that collisions are very unlikely. Serial numbers can be up to 20 octets.
func newSerialNumber() (*big.Int, error) {
	n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 8*20))
	if err != nil {
		return nil, fmt.Errorf("could not create random serial number: %s", err)
	}
	return n, nil
}
