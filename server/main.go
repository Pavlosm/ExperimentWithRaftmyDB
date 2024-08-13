package main

import (
	"flag"
	"fmt"
	"log"
	"myDb/server/cfg"
	"myDb/server/rpc"
	"net"
	"os"

	"google.golang.org/grpc"
)

var sc cfg.ServerConfig

func main() {

	m := map[string]cfg.ServerIdentity{
		"1": {
			Id:          "1",
			Port:        1551,
			BaseAddress: "localhost",
		},
		"2": {
			Id:          "2",
			Port:        1552,
			BaseAddress: "localhost",
		},
	}

	sc = cfg.ServerConfig{
		Servers: []cfg.ServerIdentity{},
	}

	myIndex := os.Args[1]

	for k, si := range m {
		if k == myIndex {
			sc.Me = si
		} else {
			sc.Servers = append(sc.Servers, si)
		}
	}

	fmt.Println(sc)
	go startServer(int32(sc.Me.Port))
	//startPing(sc)
}

// func startPing(sc state.StateMachine) {
// 	am := make(map[string]*string)
// 	for {
// 		time.Sleep(8 * time.Second)
// 		for _, server := range sc.Servers {
// 			fmt.Println("abut to ping ", server.GetUrl())

// 			a := rpc.AppendEntriesRequest{
// 				Term:     sc.ServerVars.CurrentTerm,
// 				LeaderId: sc.Cfg.Me.Id,
// 			}

// 			addr, ok := am[server.Id]
// 			if !ok {
// 				addr = flag.String("addr"+server.Id, server.GetUrl(), "the server to send")
// 				am[server.Id] = addr
// 			}

// 			conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
// 			if err != nil {
// 				fmt.Println("did not connect:", err)
// 			}
// 			defer conn.Close()
// 			c := rpc.NewRaftServiceClient(conn)

// 			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 			defer cancel()

// 			r, err := c.AppendEntries(ctx, &a)
// 			if err != nil {
// 				fmt.Println("Could not contact server", server.Id, "with error:", err)
// 			}
// 			fmt.Println(r)
// 		}
// 	}
// }

func startServer(p int32) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	rpc.RegisterRaftServiceServer(s, &rpc.RpcServer{})
	fmt.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
