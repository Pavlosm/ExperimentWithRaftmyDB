package rpcClient

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"myDb/rpc"
	"myDb/server/cfg"
	"myDb/server/logging"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConcreteRpcClient struct {
	Clients map[cfg.NodeId]rpc.RaftServiceClient
}

type RpcClient interface {
	SendAppendEntries(a *rpc.AppendEntriesRequest, reply chan<- *rpc.AppendEntriesResponse, errC chan<- error)
	SendVoteRequests(a *rpc.RequestVoteRequest, reply chan<- *rpc.RequestVoteResponse, errC chan<- error)
}

func NewRpcClient(sc cfg.ServerConfig) RpcClient {
	c := make(map[cfg.NodeId]rpc.RaftServiceClient)

	for _, server := range sc.Servers {

		if server.Id.IsEqual(sc.Me.Id) {
			continue
		}

		addr := flag.String("addr"+string(server.Id), server.GetUrl(), "the server to send")

		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			slog.Error("RPC Client: did not connect", "error", err)
		}

		c[server.Id] = rpc.NewRaftServiceClient(conn)
	}

	for k := range c {
		log.Println("Created client for", k)
	}

	return &ConcreteRpcClient{
		Clients: c,
	}
}

func (r *ConcreteRpcClient) SendAppendEntries(a *rpc.AppendEntriesRequest, reply chan<- *rpc.AppendEntriesResponse, errC chan<- error) {
	for id, c := range r.Clients {
		go func() {

			var r *rpc.AppendEntriesResponse
			var err error
			for {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				r, err = c.AppendEntries(ctx, a)
				if err == nil {
					break
				}

				slog.Error("RPC Client: SendAppendEntries could not contact server", "error", err, logging.To, string(id))
				time.Sleep(time.Second)
			}

			reply <- r
		}()
	}
}

func (r *ConcreteRpcClient) SendVoteRequests(a *rpc.RequestVoteRequest, reply chan<- *rpc.RequestVoteResponse, errC chan<- error) {
	for id, c := range r.Clients {
		go func() {

			var r *rpc.RequestVoteResponse
			var err error
			for {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				r, err = c.RequestVote(ctx, a)

				if err == nil {
					break
				}

				slog.Error("RPC Client: SendVoteRequests could not contact server waiting some seconds", "error", err, logging.To, string(id))
				time.Sleep(5 * time.Second)
			}

			reply <- r
		}()
	}
}
