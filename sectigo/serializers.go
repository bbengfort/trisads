package sectigo

// AuthenticationRequest to POST data to the authenticateEP
type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthenticationReply received from both Authenticate and Refresh
type AuthenticationReply struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// LicensesUsedResponse received from devicesEP
type LicensesUsedResponse struct {
	Ordered int `json:"ordered"`
	Issued  int `json:"issued"`
}

// AuthorityResponse received from userAuthoritiesEP
type AuthorityResponse struct {
	ID                  int    `json:"id"`
	EcosystemID         int    `json:"ecosystemId"`
	SignerCertificateID int    `json:"signerCertificateId"`
	EcosystemName       string `json:"ecosystemName"`
	Balance             int    `json:"balance"`
	Enabled             bool   `json:"enabled"`
	ProfileID           int    `json:"profileId"`
	ProfileName         string `json:"profileName"`
}

// ProfileResponse received from profilesEP
type ProfileResponse struct {
	ProfileID  int      `json:"profileId"`
	Algorithms []string `json:"algorithms"`
	CA         string   `json:"ca"`
}

// FindCertificateRequest to POST to the findCertificateEP
type FindCertificateRequest struct {
	CommonName   string `json:"commonName,omitempty"`
	SerialNumber string `json:"serialNumber,omitempty"`
}

// FindCertificateResponse from the findCertificateEP
type FindCertificateResponse struct {
	TotalCount int `json:"totalCount"`
	Items      []struct {
		DeviceID     int    `json:"deviceId"`
		CommonName   string `json:"commonName"`
		SerialNumber string `json:"serialNumber"`
		CreationDate string `json:"creationDate"`
		Status       string `json:"status"`
	} `json:"items"`
}

// RevokeCertificateRequest to POST to the revokeCertificateEP
type RevokeCertificateRequest struct {
	ReasonCode   int    `json:"reasonCode"`   // Must be code from RFC 5280 between 0 and 10
	SerialNumber string `json:"serialNumber"` // Serial number of certificated signed by profile
}
