package sectigo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shibukawa/configdir"
	"github.com/stretchr/testify/require"
)

var (
	testAccessToken  = "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIvYWNjb3VudC80Mi91c2VyLzQyIiwic2NvcGVzIjpbIlJPTEVfVVNFUiJdLCJmaXJzdC1sb2dpbiI6ZmFsc2UsImlzcyI6Imh0dHBzOi8vaW90LnNlY3RpZ28uY29tLyIsImlhdCI6MTYwNjU3MjE4OSwiZXhwIjoxNjA2NTczMDg5fQ.opGnWDYXU_1AlGCSMaZORVO7BKVR-0z5fXlsQUYhlcXxZX0Ma1kXgzUJou218iTX7pFB_38pMA6UUyE3Lpz2XQ"
	testRefreshToken = "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIvYWNjb3VudC80Mi91c2VyLzQyIiwic2NvcGVzIjpbIlJPTEVfUkVGUkVTSF9UT0tFTiJdLCJpc3MiOiJodHRwczovL2lvdC5zZWN0aWdvLmNvbS8iLCJqdGkiOiI1YTdjOThkYS05ZjA2LTRhMzYtOTBiNy04YmNhYmEwOTFlMTMiLCJpYXQiOjE2MDY1NzIxODksImV4cCI6MTYwNjU3OTM4OX0.9gV4WT1lxbitIhgD0vwyst6eF5XVs4MjIM33fpKbAUddzH6wgMgugKjC9i1ByX-P3lx0I0Zz7r3NOC3sgwzY7g"
)

func TestCredentials(t *testing.T) {
	// Skip test if a cache file already exists
	if err := checkCache(); err != nil {
		t.Skipf(err.Error())
	}

	// Ensure the environment is setup for the test and cleaned up afterward
	os.Setenv(UsernameEnv, "foo")
	os.Setenv(PasswordEnv, "secretz")
	defer os.Clearenv()

	// Load credentials from the environment with no cache
	creds := new(Credentials)
	require.NoError(t, creds.Load("", ""))

	require.Equal(t, "foo", creds.Username)
	require.Equal(t, "secretz", creds.Password)
	require.Zero(t, creds.AccessToken)
	require.Zero(t, creds.RefreshToken)

	// Set the access and refresh tokens
	require.NoError(t, creds.Update(testAccessToken, testRefreshToken))
	require.NotZero(t, creds.AccessToken)
	require.NotZero(t, creds.RefreshToken)
	require.NotZero(t, creds.Subject)
	require.NotZero(t, creds.IssuedAt)
	require.NotZero(t, creds.ExpiresAt)
	require.NotZero(t, creds.NotBefore)
	require.NotZero(t, creds.RefreshBy)

	// Load credentials from user supplied values and cached tokens
	require.NoError(t, creds.Load("teller", "tigerpaw"))

	require.Equal(t, "teller", creds.Username)
	require.Equal(t, "tigerpaw", creds.Password)
}

func checkCache() (err error) {
	cdir := configdir.New("trisa", "sectigo").QueryCacheFolder()
	if cdir.Exists("credentials.yaml") {
		return fmt.Errorf("credentials already exists at %s", filepath.Join(cdir.Path, "credentials.yaml"))
	}
	return nil
}
