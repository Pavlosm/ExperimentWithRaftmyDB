package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"myDb/server/cfg"
	"myDb/server/rpc"
	"myDb/server/rpcServer"
	"myDb/server/state"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	m := make(map[cfg.NodeId]cfg.ServerIdentity)

	for i := 1; i <= 3; i++ {
		sId := cfg.NodeId(strconv.Itoa(i))
		m[sId] = cfg.ServerIdentity{
			Id:          sId,
			Port:        1550 + i,
			BaseAddress: "localhost",
		}
	}

	myId := cfg.NodeId(os.Args[1])

	sc := cfg.NewServerConfig(myId, m)

	s := NewRpcServer(sc)

	go startServer(s, int32(sc.Me.Port))

	startPing(sc, s)
}

func startPing(sc cfg.ServerConfig, s *rpcServer.RpcServer) {
	am := make(map[cfg.NodeId]*string)
	for {
		time.Sleep(8 * time.Second)
		for _, server := range sc.Servers {

			fmt.Println("abut to ping ", server.GetUrl())

			i, t := s.State.LogState.LastLogProps()
			a := rpc.RequestVoteRequest{
				Term:         s.State.ServerVars.CurrentTerm,
				CandidateId:  string(sc.Me.Id),
				LastLogIndex: i,
				LastLogTerm:  t,
			}

			addr, ok := am[server.Id]
			if !ok {
				addr = flag.String("addr"+string(server.Id), server.GetUrl(), "the server to send")
				am[server.Id] = addr
			}

			conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Println("did not connect:", err)
			}
			defer conn.Close()
			c := rpc.NewRaftServiceClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			r, err := c.RequestVote(ctx, &a)
			if err != nil {
				fmt.Println("Could not contact server", server.Id, "with error:", err)
			}
			fmt.Println(r)
		}
	}
}

func startServer(rs *rpcServer.RpcServer, p int32) {
	flag.Parse()
	fmt.Println(p)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	rpc.RegisterRaftServiceServer(s, rs)
	fmt.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}

func NewRpcServer(c cfg.ServerConfig) *rpcServer.RpcServer {
	sv := make([]cfg.NodeId, len(c.Servers))
	for _, id := range c.Servers {
		sv = append(sv, id.Id)
	}
	s := state.NewDefaultStateMachine(sv)
	s.ServerConfig = c
	return &rpcServer.RpcServer{
		State: s,
	}
}
