/*
Package sectigo provides API access to the Sectigo IoT Manager 20.7, which is used to
sign certificate requests for directory service certificate issuance.
*/
package sectigo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Sectigo provides authenticated http requests to the Sectigo IoT Manager 20.7 REST API.
// See documentation at: https://support.sectigo.com/Com_KnowledgeDetailPage?Id=kA01N000000bvCJ
//
// Most endpoints require an JWT access token set in an Authorization: Bearer header to
// provide information about an authenticated user. The authenticate method will request
// access and refresh tokens based on user credentials. Each access token has a validity
// of 600 seconds, when the access token expires, the refresh token should be used to
// request a new access token without requiring the user to resupply credentials.
//
// The client handles authentication by checking if the tokens are valid before every
// request, and if not either refreshes the token or reauthenticates using its
// credentials.
type Sectigo struct {
	client http.Client
	creds  *Credentials
}

// New creates a Sectigo client ready to make HTTP requests, but unauthenticated. The
// username and password will be loaded from the environment if not given - from
// $SECTIGO_USERNAME and $SECTIGO_PASSWORD respectively; alternatively if not given and
// not stored in the environment, as long as valid access credentials are cached the
// credentials will be loaded.
func New(username, password string) (client *Sectigo, err error) {
	client = &Sectigo{
		creds: &Credentials{},
		client: http.Client{
			CheckRedirect: certificateAuthRedirectPolicy,
		},
	}

	if err = client.creds.Load(username, password); err != nil {
		return nil, err
	}

	return client, nil
}

// Authenticate the user with the specified credentials to get new access and refresh tokens.
// This method will replace the access tokens even if already present and valid. If
// certificate authentication is enabled then the response will be a 307 status code,
// if wrong user name and password a 401 status code and if a correct user name and
// password but the user does not have authority, a 403 status code.
func (s *Sectigo) Authenticate() (err error) {
	data := AuthenticationRequest{
		Username: s.creds.Username,
		Password: s.creds.Password,
	}

	body := new(bytes.Buffer)
	if err = json.NewEncoder(body).Encode(data); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, urlFor(authenticateEP), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	rep, err := s.client.Do(req)
	if err != nil {
		return err
	}

	// Handle error states
	switch rep.StatusCode {
	case http.StatusUnauthorized:
		return ErrInvalidCredentials
	case http.StatusForbidden:
		return ErrNotAuthorized
	case http.StatusTemporaryRedirect:
		return ErrMustUseTLSAuth
	}

	if rep.StatusCode != http.StatusOK {
		return fmt.Errorf("unhandled status code: %d", rep.StatusCode)
	}

	// We've got a successful response - deserialize request body
	tokens := &AuthenticationReply{}
	if err = json.NewDecoder(rep.Body).Decode(&tokens); err != nil {
		return err
	}

	if err = s.creds.Update(tokens.AccessToken, tokens.RefreshToken); err != nil {
		return err
	}
	return nil
}

// Refresh the access token using the refresh token. Note that this method does not
// check if the credentials are refreshable, it only issues the refresh request with
// the refresh access token if it exists. If the refresh token does not exist, then an
// error is returned.
func (s *Sectigo) Refresh() (err error) {
	if s.creds.RefreshToken == "" {
		return ErrNotAuthenticated
	}

	body := new(bytes.Buffer)
	fmt.Fprintf(body, "%s\n", s.creds.RefreshToken)

	req, err := http.NewRequest(http.MethodPost, urlFor(refreshEP), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	rep, err := s.client.Do(req)
	if err != nil {
		return err
	}

	// Handle error states
	switch rep.StatusCode {
	case http.StatusUnauthorized:
		return ErrInvalidCredentials
	case http.StatusForbidden:
		return ErrNotAuthorized
	}

	if rep.StatusCode != http.StatusOK {
		return fmt.Errorf("unhandled status code: %d", rep.StatusCode)
	}

	// We've got a successful response - deserialize request body
	tokens := &AuthenticationReply{}
	if err = json.NewDecoder(rep.Body).Decode(&tokens); err != nil {
		return err
	}

	// It appears that sectigo reuses the refresh token.
	// TODO: verify refresh behavior to ensure that it's used correctly
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = s.creds.RefreshToken
	}

	if err = s.creds.Update(tokens.AccessToken, tokens.RefreshToken); err != nil {
		return err
	}
	return nil
}

// LicensesUsed returns statistic for Ordered/Issued certificates (licenses used)
// User must be authenticated with role 'USER'
func (s *Sectigo) LicensesUsed() (stats *LicensesUsedResponse, err error) {
	// perform preflight check for authenticated endpoint
	if err = s.preflight(); err != nil {
		return nil, err
	}

	// create request
	var req *http.Request
	if req, err = s.newRequest(http.MethodGet, urlFor(devicesEP), nil); err != nil {
		return nil, err
	}

	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return nil, err
	}

	if err = s.checkStatus(rep); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(rep.Body).Decode(&stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// UserAuthorities returns a list of all Authorities by Ecosystem and Current User
// User must be authenticated.
func (s *Sectigo) UserAuthorities() (authorities []*AuthorityResponse, err error) {
	// perform preflight check for authenticated endpoint
	if err = s.preflight(); err != nil {
		return nil, err
	}

	// create request
	var req *http.Request
	if req, err = s.newRequest(http.MethodGet, urlFor(userAuthoritiesEP), nil); err != nil {
		return nil, err
	}

	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return nil, err
	}

	if err = s.checkStatus(rep); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(rep.Body).Decode(&authorities); err != nil {
		return nil, err
	}
	return authorities, nil
}

// Profiles returns a list of all profiles available to the user.
// User must be authenticated.
func (s *Sectigo) Profiles() (profiles []*ProfileResponse, err error) {
	// perform preflight check for authenticated endpoint
	if err = s.preflight(); err != nil {
		return nil, err
	}

	// create request
	var req *http.Request
	if req, err = s.newRequest(http.MethodGet, urlFor(profilesEP), nil); err != nil {
		return nil, err
	}

	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return nil, err
	}

	if err = s.checkStatus(rep); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(rep.Body).Decode(&profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

// FindCertificate searches for certificates by common name and serial number.
func (s *Sectigo) FindCertificate(commonName, serialNumber string) (certs *FindCertificateResponse, err error) {
	// perform preflight check for authenticated endpoint
	if err = s.preflight(); err != nil {
		return nil, err
	}

	query := &FindCertificateRequest{
		CommonName:   commonName,
		SerialNumber: serialNumber,
	}

	// create request
	var req *http.Request
	if req, err = s.newRequest(http.MethodPost, urlFor(findCertificateEP), query); err != nil {
		return nil, err
	}

	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return nil, err
	}

	if err = s.checkStatus(rep); err != nil {
		return nil, err
	}

	if err = json.NewDecoder(rep.Body).Decode(&certs); err != nil {
		return nil, err
	}
	return certs, nil
}

// RevokeCertificate by serial number if the certificate was signed by the given authority.
// A reason code from RFC 5280 must be given. This method revokes single certificates
// unlike the RevokeDeviceCertificates method which can revoke multiple certificates by
// their assignment to specific Device IDs. If no error is returned, the certificate
// revocation was successful.
// User must be authenticated and has permission to update profile.
func (s *Sectigo) RevokeCertificate(profileID, reasonCode int, serialNumber string) (err error) {
	// perform preflight check for authenticated endpoint
	if err = s.preflight(); err != nil {
		return err
	}

	query := &RevokeCertificateRequest{
		ReasonCode:   reasonCode,
		SerialNumber: serialNumber,
	}

	// create request
	var req *http.Request
	if req, err = s.newRequest(http.MethodPost, urlFor(revokeCertificateEP, profileID), query); err != nil {
		return err
	}

	var rep *http.Response
	if rep, err = s.client.Do(req); err != nil {
		return err
	}

	if err = s.checkStatus(rep); err != nil {
		return err
	}
	return nil
}

// Creds returns a copy of the underlying credentials object.
func (s *Sectigo) Creds() Credentials {
	return *s.creds
}

// Returns a request with default headers set along with the authentication header.
// If the client has not been authenticated, then an error is returned.
func (s *Sectigo) newRequest(method, url string, data interface{}) (req *http.Request, err error) {
	if !s.creds.Valid() {
		return nil, ErrNotAuthenticated
	}

	// JSON serialize the data being sent to the request
	var body io.Reader
	if data != nil {
		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(data)
	}

	// Create the request
	if req, err = http.NewRequest(method, url, body); err != nil {
		return nil, err
	}

	// Set Headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.creds.AccessToken))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	return req, nil
}

// Preflight prepares to send a request that needs to be authenticated by checking the
// credentials and sending any authentication or refresh requests required.
func (s *Sectigo) preflight() (err error) {
	if !s.creds.Valid() {
		if s.creds.Refreshable() {
			// Attempt to refresh the credentials, if there is no error, then continue.
			// However, the refresh endpoint may not be working so if it errors, attempt
			// to reauthenticate with username and password instead.
			// TODO: add logging so the server knows what's going on.
			if err = s.Refresh(); err == nil {
				return err
			}
		}

		// If we could not refresh, attempt to reauthenticate
		if err = s.Authenticate(); err != nil {
			return err
		}
	}

	// Check the credentials and if they're good, dump them to disk
	if err = s.creds.Check(); err != nil {
		s.creds.Clear()
		s.creds.Dump()
		return err
	}

	// Ignore any cache errors on reauthentication
	s.creds.Dump()

	// Good to go!
	return nil
}

// Helper function to convert a non-200 HTTP status into an error, reading JSON error
// data if it's available, otherwise returning a simple error. Note that this method
// will attempt to read the body on error, so do not use it for error handling that
// requires knowledge of the body.
func (s *Sectigo) checkStatus(rep *http.Response) (err error) {
	// Check if status code is a good status.
	if rep.StatusCode >= 200 && rep.StatusCode < 300 {
		return nil
	}

	// Try to unmarshall the error from the response
	var e *APIError
	if err = json.NewDecoder(rep.Body).Decode(&e); err != nil {
		switch rep.StatusCode {
		case http.StatusUnauthorized:
			return ErrNotAuthenticated
		case http.StatusForbidden:
			return ErrNotAuthorized
		}

		// Return a simple error since the JSON could not be decoded.
		e = &APIError{
			Status:  rep.StatusCode,
			Message: rep.Status,
		}
	}
	return e
}

// The Sectigo API has a special authentication policy when certificate authentication
// is enabled. In this case, normal password authentication requests return a 307 with
// a URL to POST to the certificate auth location, meaning that the URL requires TLS
// client authentication. TLS client authentication is only required to obtain an access
// token. This function prevents multiple redirects in the case of a 307 by returning
// the redirect request. Other redirect status codes are followed.
func certificateAuthRedirectPolicy(req *http.Request, via []*http.Request) error {
	if req.Response.StatusCode == http.StatusTemporaryRedirect {
		return http.ErrUseLastResponse
	}
	return nil
}
