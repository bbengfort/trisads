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

// CreateSingleCertBatchRequest to POST data to the createSingleCertBatchEP
type CreateSingleCertBatchRequest struct {
	AuthorityID   int               `json:"authorityId"`
	BatchName     string            `json:"batchName"`
	ProfileParams map[string]string `json:"profileParams"` // should not be empty; represents the profile-specific params passed to batch request
}

// BatchResponse received from createSingleCertBatchEP and batchDetailEP
type BatchResponse struct {
	BatchID         int         `json:"batchId"`
	OrderNumber     int         `json:"orderNumber"`
	CreationDate    string      `json:"creationDate"`
	Profile         string      `json:"profile"`
	Size            int         `json:"size"`
	Status          string      `json:"status"`
	Active          bool        `json:"active"`
	BatchName       string      `json:"batchName"`
	RejectReason    string      `json:"rejectReason"`
	GeneratorValues interface{} `json:"generatorParametersValues"`
	UserID          int         `json:"userId"`
	Downloadable    bool        `json:"downloadable"`
	Rejectable      bool        `json:"rejectable"`
}

// ProcessingInfoResponse received from batchProcessingInfoEP
type ProcessingInfoResponse struct {
	Active  int `json:"active"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
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

// ProfileParamsResponse received from profileParametersEP
type ProfileParamsResponse struct {
	Name              string      `json:"name"`
	InputType         string      `json:"inputType"`
	Required          bool        `json:"required"`
	Placeholder       interface{} `json:"placeholder"`
	ValidationPattern string      `json:"validationPattern"`
	Message           string      `json:"message"`
	Value             interface{} `json:"value"`
	Title             string      `json:"title"`
	Scopes            []string    `json:"scopes"`
	Dynamic           bool        `json:"dynamic"`
}

// ProfileDetailResponse received from profileDetailEP
type ProfileDetailResponse struct {
	ProfileName      string `json:"profileName"`
	ProfileID        int    `json:"profileId"`
	RawProfileConfig string `json:"rawProfileConfig"`
	Name             string `json:"name"`
	KeyAlgorithmInfo string `json:"keyAlgorithmInfo"`
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
