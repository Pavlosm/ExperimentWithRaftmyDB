package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"myDb/rpc"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	c := NewRpcClient()
	c.SetLeader("1")
	for {
		fmt.Print("Enter command: ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command", err)
			continue
		}
		cmd = strings.TrimSpace(cmd)
		if cmd == "exit" {
			os.Exit(0)
		}

		for {
			r := c.SendCommand(&rpc.CommandRequest{Command: cmd})
			if !r.Success {
				if r.LeaderId != c.GetLeader() {
					c.SetLeader(r.LeaderId)
					fmt.Println("Redirecting to leader", c.GetLeader())
				} else {
					fmt.Println("Command failed, retrying after 3 seconds")
					time.Sleep(3 * time.Second)
				}
			} else {
				break
			}
		}

		fmt.Println(cmd)
	}
	// TODO: prompt to get command
	// TODO: send append entries command
	// TODO: is the server is other than leader, redirect to the leader
}

type ConcreteRpcClient struct {
	Clients  map[string]rpc.RaftServiceClient
	LeaderId string
}

type RpcClient interface {
	SendCommand(a *rpc.CommandRequest) *rpc.CommandResponse
	SetLeader(leaderId string)
	GetLeader() string
}

func (r *ConcreteRpcClient) GetLeader() string {
	return r.LeaderId
}

func (r *ConcreteRpcClient) SetLeader(leaderId string) {
	r.LeaderId = leaderId
}

func (r *ConcreteRpcClient) SendCommand(a *rpc.CommandRequest) *rpc.CommandResponse {
	c := r.Clients[r.LeaderId]
	var res *rpc.CommandResponse
	var err error
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		res, err = c.SendCommand(ctx, a)
		if err == nil {
			break
		}

		fmt.Println("RPC Client: SendCommand could not contact server", "error", err, "to", r.LeaderId)
		time.Sleep(time.Second)
	}

	return res
}

func NewRpcClient() RpcClient {
	c := make(map[string]rpc.RaftServiceClient)

	servers := map[string]string{"1": "localhost:1551", "2": "localhost:1552", "3": "localhost:1553"}
	for k, v := range servers {

		addr := flag.String("addr"+k, v, "the server to send")

		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			slog.Error("RPC Client: did not connect", "error", err)
		}

		c[k] = rpc.NewRaftServiceClient(conn)
	}

	for k := range c {
		log.Println("Created client for", k)
	}

	return &ConcreteRpcClient{
		Clients: c,
	}
}
