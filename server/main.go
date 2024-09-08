package main

import (
	"flag"
	"fmt"
	"log"
	"myDb/server/cfg"
	"myDb/server/rpc"
	"myDb/server/rpcClient"
	"myDb/server/rpcServer"
	"myDb/server/state"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

var sc cfg.ServerConfig
var c FlowControlChannels
var s rpcServer.RpcServer
var st state.StateMachineHandler
var rc rpcClient.RpcClient

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

	sc = cfg.NewServerConfig(myId, m)

	c = NewFlowControlChannels()

	server, stateMachine := NewRpcServer(sc, &c)
	s = *server
	st = stateMachine
	rc = *rpcClient.NewRpcClient(sc)

	go startServer()

	quit := make(chan int)
	<-quit
}

func startServer() {

	go startRpcServer()

	st.Start()

	for {
		select {
		case v := <-c.VoteCmd: // external command
			log.Println("MAIN: received vote request from", v.Data.ServerId)
			granted := st.HandleVoteRequest(
				v.Data.ServerId,
				v.Data.Term,
				v.Data.LastLogTerm,
				v.Data.LastLogIndex)

			v.Reply <- &rpc.RequestVoteResponse{
				ServerId:    st.GetServerIdString(),
				Term:        st.GetCurrentTerm(),
				VoteGranted: granted,
			}
			log.Println("MAIN: sent vote response", granted, "to", v.Data.ServerId)
		case a := <-c.AppendEntriesCmd: // external command
			log.Println("MAIN: received AppendEntries request from", a.Data.ServerId, "with term", a.Data.Term, ". My term is", st.GetCurrentTerm())
			st.HandleAppendEntriesRequest(a.Data.Term)
			a.Reply <- &rpc.AppendEntriesResponse{
				ServerId:   st.GetServerIdString(),
				Term:       st.GetCurrentTerm(),
				Success:    true,
				MatchIndex: 1,
			}
			log.Println("MAIN: handled AppendEntries request from", a.Data.ServerId, "with term", a.Data.Term, ". My term is", st.GetCurrentTerm())
		case m := <-c.VoteReply: // response from request command
			log.Println("New vote from", m.ServerId, "for term,", m.Term, "with value", m.VoteGranted)
			st.HandleVoteResponse(m.ServerId, m.VoteGranted, m.Term)
		case t := <-st.GetVoteFireChan(): // internal command
			log.Println("MAIN: expiry timer fired. time:", t)
			if st.HandleVoteTimerFired() {
				i, t := st.GetLastLogProps()
				rc.SendVoteRequests(&rpc.RequestVoteRequest{
					Term:         st.GetCurrentTerm(),
					ServerId:     st.GetServerIdString(),
					LastLogIndex: i,
					LastLogTerm:  t,
				}, c.VoteReply, c.Err)
			}
		case tm := <-st.GetPingFireChan():
			log.Println("MAIN: expiry timer fired. time:", tm)
			i, t := st.GetLastLogProps()
			rc.SendAppendEntries(&rpc.AppendEntriesRequest{
				Term:         st.GetCurrentTerm(),
				ServerId:     st.GetServerIdString(),
				PrevLogIndex: i,
				PrevLogTerm:  t,
				LeaderCommit: 0,
				Entries:      make([]*rpc.Entry, 0),
			}, c.AppendEntriesReply, c.Err)
			st.HandlePingRequestSent()
		case a := <-c.AppendEntriesReply:
			log.Println("Received append entries reply", a)
		case e := <-c.Err:
			log.Println("An error ocurred", e)
		}
	}
}

func startRpcServer() {
	flag.Parse()

	p := int32(sc.Me.Port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()

	rpc.RegisterRaftServiceServer(gs, &s)
	log.Printf("MAIN: server listening at %v\n", lis.Addr())

	if err := gs.Serve(lis); err != nil {
		log.Printf("MAIN: failed to serve: %v\n", err)
	}
}
