/*
Package store provides an interface to database storage for the TRISA directory service.
*/
package store

import (
	"fmt"
	"net/url"

	"github.com/bbengfort/trisads/pb"
)

// Open a directory storage provider with the specified URI. Database URLs should either
// specify protocol+transport://user:pass@host/dbname?opt1=a&opt2=b for servers or
// protocol:/path/to/file for embedded databases.
func Open(uri string) (Store, error) {
	dsn, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// If no scheme is specified, default to leveldb expecting a path
	if dsn.Scheme == "" && dsn.Path != "" {
		return OpenLevelDB(uri)
	}

	switch dsn.Scheme {
	case "leveldb":
		return OpenLevelDB(uri)
	default:
		return nil, fmt.Errorf("unhandled database scheme %q", dsn.Scheme)
	}
}

// Store provides an interface for directory storage services to abstract the underlying
// database provider. The storage methods correspond to directory service requests,
// which are currently implemented with a simple CRUD and search interface for VASP
// records. The underlying database can be a simple embedded store or a distributed
// SQL server, so long as it can interact with VASP identity records.
type Store interface {
	Close() error
	Create(v pb.VASP) (uint64, error)
	Retrieve(id uint64) (pb.VASP, error)
	Update(v pb.VASP) error
	Destroy(id uint64) error
	Search(query map[string]interface{}) ([]pb.VASP, error)
}
