package rpcservice

import (
	"github.com/capdale/was/config"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RpcService struct {
	ImageClassifies *[]*ImageClassify
}

// naming as service name
type ImageClassify struct {
	Client  *rpc_protocol.ImageClassifyClient
	Conn    *grpc.ClientConn
	IsAlive bool // not used yet, for recovering dead service
}

func New(c *config.Rpc) (*RpcService, error) {
	conns, err := createConn(&c.ImageClassify.Address)
	if err != nil {
		return nil, err
	}
	imageClassifies := createImageClassifies(conns)

	return &RpcService{
		ImageClassifies: imageClassifies,
	}, nil
}

func createConn(addresses *[]string) (*[]*grpc.ClientConn, error) {
	conns := []*grpc.ClientConn{}

	for _, addr := range *addresses {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return &conns, nil
}

func createImageClassifies(conns *[]*grpc.ClientConn) *[]*ImageClassify {
	clients := []*ImageClassify{}
	for _, conn := range *conns {
		imageClassiferClient := rpc_protocol.NewImageClassifyClient(conn)
		clients = append(clients, &ImageClassify{
			Client:  &imageClassiferClient,
			Conn:    conn,
			IsAlive: true,
		})
	}
	return &clients
}
