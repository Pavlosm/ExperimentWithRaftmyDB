package rpcClient

import (
	"context"
	"flag"
	"log"
	"myDb/server/cfg"
	"myDb/server/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RpcClient struct {
	Clients map[cfg.NodeId]rpc.RaftServiceClient
}

func NewRpcClient(sc cfg.ServerConfig) *RpcClient {
	c := make(map[cfg.NodeId]rpc.RaftServiceClient)

	for _, server := range sc.Servers {

		if server.Id.IsEqual(sc.Me.Id) {
			continue
		}

		addr := flag.String("addr"+string(server.Id), server.GetUrl(), "the server to send")

		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Println("did not connect:", err)
		}

		c[server.Id] = rpc.NewRaftServiceClient(conn)
	}

	return &RpcClient{
		Clients: c,
	}
}

func (r *RpcClient) SendAppendEntries(a *rpc.AppendEntriesRequest, reply chan<- *rpc.AppendEntriesResponse, errC chan<- error) {
	for id, c := range r.Clients {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r, err := c.AppendEntries(ctx, a)

			if err != nil {
				log.Println("Could not contact server", id, "with error:", err)
				errC <- err
				return
			}

			reply <- r
		}()
	}
}

func (r *RpcClient) SendVoteRequests(a *rpc.RequestVoteRequest, reply chan<- *rpc.RequestVoteResponse, errC chan<- error) {
	for id, c := range r.Clients {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r, err := c.RequestVote(ctx, a)

			if err != nil {
				log.Println("Could not contact server", id, "with error:", err)
				errC <- err
				return
			}

			reply <- r
		}()
	}
}
