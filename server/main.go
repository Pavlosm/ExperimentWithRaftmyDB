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
	"time"

	"google.golang.org/grpc"
)

var sc cfg.ServerConfig
var c FlowControlChannels
var s rpcServer.RpcServer
var st state.StateMachine
var tm TimeoutMod
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
	st = *stateMachine

	tm = *NewTimeoutMod()

	rc = *rpcClient.NewRpcClient(sc)

	go startServer()

	quit := make(chan int)
	<-quit
}

func startServer() {

	go startRpcServer()

	go tm.Start()

	go appendEntriesPing()

	for {
		select {
		case v := <-c.VoteCmdChan:
			log.Println("MAIN: received vote request from", v.Data.CandidateId)

			granted := st.HandleVoteRequest(
				v.Data.CandidateId,
				v.Data.Term,
				v.Data.LastLogTerm,
				v.Data.LastLogIndex)

			vr := &rpc.RequestVoteResponse{
				Term:        st.GetCurrentTerm(),
				VoteGranted: granted,
			}
			v.Reply <- vr
			if st.ServerVars.Role == state.Leader {
				tm.StopChan <- true
			}
			log.Println("MAIN: sent vote response", granted, "to", v.Data.CandidateId)
		case a := <-c.AppendEntriesCmdChan:
			log.Println("MAIN: received vote request from", a.Data.LeaderId, "with term", a.Data.Term, ". My term is", st.ServerVars.CurrentTerm)

			st.HandleAppendEntriesRequest(a.Data.Term)

			ar := &rpc.AppendEntriesResponse{
				Term:       st.GetCurrentTerm(),
				Success:    true,
				MatchIndex: 1,
			}
			a.Reply <- ar

			if st.ServerVars.Role == state.Leader {
				tm.StopChan <- true
			} else {
				tm.ResetChan <- true
			}

			log.Println("MAIN: handled AppendEntries request from", a.Data.LeaderId, "with term", a.Data.Term, ". My term is", st.ServerVars.CurrentTerm)
		case _, ok := <-tm.FireChan:
			log.Println("MAIN: expiry timer fired")
			if ok {
				st.ServerVars.Role = state.Candidate
				// TODO start voting process processing for
			}
			tm.ResetChan <- true
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

func appendEntriesPing() {
	for {
		time.Sleep(2 * time.Second)
		i, t := st.LogState.LastLogProps()

		a := rpc.AppendEntriesRequest{
			Term:         st.ServerVars.CurrentTerm,
			LeaderId:     string(sc.Me.Id),
			PrevLogIndex: i,
			PrevLogTerm:  t,
			LeaderCommit: 0,
			Entries:      make([]*rpc.Entry, 0),
		}

		rc.SendAppendEntries(&a)
	}
}
