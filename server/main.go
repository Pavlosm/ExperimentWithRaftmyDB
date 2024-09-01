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

	for i := 1; i <= 2; i++ {
		sId := cfg.NodeId(strconv.Itoa(i))
		m[sId] = cfg.ServerIdentity{
			Id:          sId,
			Port:        1550 + i,
			BaseAddress: "localhost",
		}
	}

	myId := cfg.NodeId(os.Args[1])

	fmt.Println(myId)
	sc := cfg.NewServerConfig(myId, m)

	c := make(chan int)

	req := make(chan *rpc.RequestVoteRequest)
	res := make(chan *rpc.RequestVoteResponse)
	areq := make(chan *rpc.AppendEntriesRequest)
	ares := make(chan *rpc.AppendEntriesResponse)
	s, st := NewRpcServer(sc, c, req, res)

	go startServer(s, int32(sc.Me.Port), req, res, areq, ares)

	startPing(sc, st)

	//_ := make([]string, 0, 1000)
}

func startPing(sc cfg.ServerConfig, st state.StateMachine) {
	am := make(map[cfg.NodeId]*string)
	for {
		time.Sleep(15 * time.Second)
		for _, server := range sc.Servers {

			log.Println("about to ping ", server.GetUrl())

			i, t := st.LogState.LastLogProps()
			a := rpc.RequestVoteRequest{
				Term:         st.ServerVars.CurrentTerm,
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

func startServer(
	rs *rpcServer.RpcServer,
	p int32,
	r chan *rpc.RequestVoteRequest,
	vres chan *rpc.RequestVoteResponse,
	a chan *rpc.AppendEntriesRequest,
	ares chan *rpc.AppendEntriesResponse) {

	go func() {
		flag.Parse()

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
	}()

	for {
		select {
		case in := <-r:
			log.Println("MAIN: received request vote request")
			granted := rs.State.HandleVoteRequest(
				in.CandidateId,
				in.Term,
				in.LastLogTerm,
				in.LastLogIndex)
			log.Println("MAIN: handled request vote request")
			resp := &rpc.RequestVoteResponse{
				Term:        rs.State.GetCurrentTerm(),
				VoteGranted: granted,
			}
			vres <- resp
			log.Println("MAIN: sent to response channel")
		case in := <-a:
			fmt.Println("Not supported", in)
		}
	}
}
