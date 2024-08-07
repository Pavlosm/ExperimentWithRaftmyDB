package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"myDb/server/cfg"
	"myDb/server/rpc"
	"myDb/server/stateMachine"
	"net"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	s := [...]cfg.ServerIdentity{
		{
			Id:          "1",
			Port:        1551,
			BaseAddress: "localhost",
		},
		{
			Id:          "2",
			Port:        1552,
			BaseAddress: "localhost",
		},
		{
			Id:          "3",
			Port:        1553,
			BaseAddress: "localhost",
		},
	}

	sc := stateMachine.StateMachine{
		Term: 0,
		Cfg: cfg.ServerConfig{
			Servers: []cfg.ServerIdentity{},
		},
		Role: cfg.Follower,
	}

	myIndex, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic(err)
	}

	for i, si := range s {
		if i == myIndex {
			sc.Cfg.Me = si
		} else {
			sc.Cfg.Servers = append(sc.Cfg.Servers, si)
		}
	}

	fmt.Println(sc)
	go startServer(int32(sc.Cfg.Me.Port))
	startPing(sc)
}

func startPing(sc stateMachine.StateMachine) {
	am := make(map[string]*string)
	for {
		time.Sleep(1 * time.Second)
		for _, server := range sc.Cfg.Servers {
			fmt.Println("Ping ", server.GetUrl())

			a := rpc.AppendEntriesRequest{
				Term: sc.Term,
				Id:   sc.Cfg.Me.Id,
			}

			addr, ok := am[server.Id]
			if !ok {
				addr = flag.String("addr"+server.Id, server.GetUrl(), "the server to send")
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

			r, err := c.SendAppendEntries(ctx, &a)
			if err != nil {
				fmt.Println("Could not contact server", server.Id, "with error:", err)
			}
			fmt.Println(r)
		}
	}
}

type server struct {
	rpc.UnimplementedRaftServiceServer
}

func startServer(p int32) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	rpc.RegisterRaftServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
