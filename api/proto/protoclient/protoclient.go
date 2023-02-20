// Package protoclient provides a client to connect to nubedb's gRPC server.
package protoclient

import (
	"context"
	"github.com/narvikd/errorskit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"nubedb/api/proto"
	"time"
)

// ConnCleaner is an interface for cleaning up connections.
type ConnCleaner interface {
	Cleanup()
}

// Connection represents a connection to a gRPC server.
type Connection struct {
	Conn          *grpc.ClientConn
	Client        proto.ServiceClient
	Ctx           context.Context
	cancelCtxCall context.CancelFunc
}

// Cleanup closes the connection and cancels the context.
func (c *Connection) Cleanup() {
	_ = c.Conn.Close()
	c.cancelCtxCall()
}

// NewConnection creates a new connection to a gRPC server.
//
// Example:
//
//	conn, errConn := protoclient.NewConnection(leaderGrpcAddr)
//	if errConn != nil {
//		return errConn
//	}
//	defer conn.Cleanup()
func NewConnection(addr string) (*Connection, error) {
	const (
		dialTimeout       = 1 * time.Second
		timeoutGrpcCall   = 3 * time.Second
		errGrpcConnection = "grpc connection failed"
	)

	// Create a context with a timeout for dialing the server.
	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelCtxDial()

	// Dial the server using the context.
	connDial, errDial := grpc.DialContext(ctxDial, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return nil, errorskit.Wrap(errDial, errGrpcConnection)
	}

	// Create a context with a timeout for executing gRPC calls.
	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)

	return &Connection{
		Conn:          connDial,                         // set the gRPC connection
		Client:        proto.NewServiceClient(connDial), // create a new service client
		Ctx:           ctxExecuteCall,                   // set the context for executing gRPC calls
		cancelCtxCall: cancelCtxExecuteCall,             // set the cancel function for the context included before
	}, nil
}
