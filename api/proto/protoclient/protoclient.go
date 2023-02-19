package protoclient

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"nubedb/api/proto"
	"nubedb/pkg/errorskit"
	"time"
)

type CleanFunc func()

type ConnCleaner interface {
	Cleanup()
}

type Connection struct {
	Conn          *grpc.ClientConn
	Client        proto.ServiceClient
	Ctx           context.Context
	cancelCtxCall context.CancelFunc
}

func (c *Connection) Cleanup() {
	_ = c.Conn.Close()
	c.cancelCtxCall()
}

func NewConnection(addr string) (*Connection, error) {
	const (
		dialTimeout       = 1 * time.Second
		timeoutGrpcCall   = 3 * time.Second
		errGrpcConnection = "grpc connection failed"
	)

	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), dialTimeout)
	defer cancelCtxDial()

	connDial, errDial := grpc.DialContext(ctxDial, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return nil, errorskit.Wrap(errDial, errGrpcConnection)
	}

	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)

	return &Connection{
		Conn:          connDial,
		Client:        proto.NewServiceClient(connDial),
		Ctx:           ctxExecuteCall,
		cancelCtxCall: cancelCtxExecuteCall,
	}, nil
}
