package rpcclient

import (
	"github.com/capdale/was/config"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RpcClient struct {
	ImageClassifyClient *rpc_protocol.ImageClassifyClient
}

func New(c *config.Rpc) (*grpc.ClientConn, *RpcClient, error) {
	conn, err := grpc.Dial(c.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	imageClassiferClient := rpc_protocol.NewImageClassifyClient(conn)
	return conn, &RpcClient{
		ImageClassifyClient: &imageClassiferClient,
	}, nil
}
