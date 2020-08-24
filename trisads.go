package trisads

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/bbengfort/trisads/pb"
	"github.com/bbengfort/trisads/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// New creates a TRISA Directory Service with the specified configuration and prepares
// it to listen for and serve GRPC requests.
func New(dsn string) (s *Server, err error) {
	s = &Server{}
	if s.db, err = store.Open(dsn); err != nil {
		return nil, err
	}

	return s, nil
}

// Server implements the GRPC TRISADirectoryService.
type Server struct {
	db  store.Store
	srv *grpc.Server
}

// Serve GRPC requests on the specified address.
func (s *Server) Serve(addr string) (err error) {
	// Initialize the gRPC server
	s.srv = grpc.NewServer()
	pb.RegisterTRISADirectoryServer(s.srv, s)

	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.Shutdown()
	}()

	// Listen for TCP requests on the specified address and port
	var sock net.Listener
	if sock, err = net.Listen("tcp", addr); err != nil {
		return fmt.Errorf("could not listen on %q", addr)
	}
	defer sock.Close()

	// Run the server
	log.Infof("listening on %s", addr)
	return s.srv.Serve(sock)
}

// Shutdown the TRISA Directory Service gracefully
func (s *Server) Shutdown() (err error) {
	log.Info("gracefully shutting down")
	s.srv.GracefulStop()
	if err = s.db.Close(); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

// Register a new VASP entity with the directory service. After registration, the new
// entity must go through the verification process to get issued a certificate. The
// status of verification can be obtained by using the lookup RPC call.
func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (out *pb.RegisterReply, err error) {
	out = &pb.RegisterReply{}
	vasp := pb.VASP{VaspEntity: in.Entity}

	if out.Id, err = s.db.Create(vasp); err != nil {
		log.WithError(err).Warn("could not register VASP")
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
	} else {
		log.WithField("name", in.Entity.VaspFullLegalName).Info("registered VASP")
	}

	// TODO: if verify is true: send verification request
	return out, nil
}

// Lookup a VASP entity by name or ID to get full details including the TRISA certification
// if it exists and the entity has been verified.
func (s *Server) Lookup(ctx context.Context, in *pb.LookupRequest) (out *pb.LookupReply, err error) {
	var vasp pb.VASP
	out = &pb.LookupReply{}

	if in.Id > 0 {
		if vasp, err = s.db.Retrieve(in.Id); err != nil {
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

	} else if in.Name != "" {
		var vasps []pb.VASP
		if vasps, err = s.db.Search(map[string]interface{}{"name": in.Name}); err != nil {
			out.Error = &pb.Error{
				Code:    404,
				Message: err.Error(),
			}
		}

		if len(vasps) == 1 {
			vasp = vasps[0]
		} else {
			out.Error = &pb.Error{
				Code:    404,
				Message: "not found",
			}
		}
	} else {
		out.Error = &pb.Error{
			Code:    400,
			Message: "no lookup query provided",
		}
		return out, nil
	}

	if out.Error == nil {
		out.Vasp = &vasp
		log.WithField("id", vasp.Id).Info("VASP lookup succeeded")
	} else {
		log.WithError(out.Error).Warn("could not lookup VASP")
	}
	return out, nil
}

// Search for VASP entity records by name or by country in order to perform more detailed
// Lookup requests. The search process is purposefully simplistic at the moment.
func (s *Server) Search(ctx context.Context, in *pb.SearchRequest) (out *pb.SearchReply, err error) {
	out = &pb.SearchReply{}
	query := make(map[string]interface{})
	query["name"] = in.Name
	query["country"] = in.Country

	var vasps []pb.VASP
	if vasps, err = s.db.Search(query); err != nil {
		out.Error = &pb.Error{
			Code:    400,
			Message: err.Error(),
		}
	}

	out.Vasps = make([]*pb.VASP, len(vasps))
	for i, vasp := range vasps {
		// return only entities, remove certificate info until lookup
		vasp.VaspTRISACertification = nil
		out.Vasps[i] = &vasp
	}

	entry := log.WithFields(log.Fields{
		"name":    in.Name,
		"country": in.Country,
		"results": len(out.Vasps),
	})
	if out.Error != nil {
		entry.WithError(out.Error).Warn("unsuccessful search")
	} else {
		entry.Info("search succeeded")
	}
	return out, nil
}
