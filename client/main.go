package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"myDb/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	servers := map[string]string{"1": "localhost:1551", "2": "localhost:1552", "3": "localhost:1553"}
	c := NewRpcClient(servers)
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

		var errCount int
		for {
			r := c.SendCommand(&rpc.CommandRequest{Command: cmd})

			if r == nil && errCount > 2 {
				targetLeaderInt := rand.Intn(len(servers)) + 1
				targetLeader := strconv.Itoa(targetLeaderInt)
				fmt.Println("Reached max number of retries, redirecting to another server. Current leader", c.GetLeader(), "Target leader", targetLeader)
				c.SetLeader(targetLeader)
				errCount = 0
			} else if r == nil {
				fmt.Println("Command failed, retrying after 1 second")
				time.Sleep(time.Second)
				errCount++
			} else if !r.Success {
				l := c.GetLeader()
				if r.LeaderId != l && l != "" {
					c.SetLeader(r.LeaderId)
					errCount = 0
					fmt.Println("Redirecting to leader", l)
				} else {
					fmt.Println("Command failed. No retry will be attempted")
					break
				}
			} else {
				break
			}
		}
	}
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

	if c == nil {
		fmt.Println("RPC Client: could not find the appropriate client")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := c.SendCommand(ctx, a)

	if err != nil {
		return nil
	}

	return res
}

func NewRpcClient(servers map[string]string) RpcClient {
	c := make(map[string]rpc.RaftServiceClient)

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
