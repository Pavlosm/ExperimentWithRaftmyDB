package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"myDb/server/cfg"
	"myDb/server/environment"
	"myDb/server/logging"
	"myDb/server/rpc"
	"myDb/server/rpcClient"
	"myDb/server/rpcServer"
	"myDb/server/state"
	"myDb/server/utils"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
)

var sc cfg.ServerConfig
var e *environment.Env
var s rpcServer.RpcServer
var st state.StateMachineHandler
var rc rpcClient.RpcClient

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

	fmt.Println(myId)

	sc = cfg.NewServerConfig(myId, m)

	e = environment.NewEnvironment()

	server, stateMachine := NewRpcServer(sc, e)
	s = *server
	st = stateMachine
	rc = rpcClient.NewRpcClient(sc)

	go startServer()

	quit := make(chan int)
	<-quit
}

func startServer() {

	go startRpcServer()

	st.Start()

	for {
		select {
		case v := <-e.Channels.VoteCmd: // external command
			info("received vote request", v.Data.ServerId, "", "")

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

			info("sent vote response", "", v.Data.ServerId, "")
		case a := <-e.Channels.AppendEntriesCmd: // external command
			slog.Info("received AppendEntries request", "from", a.Data.ServerId, "myTerm", st.GetCurrentTerm(), "mTerm", a.Data.Term)
			st.HandleAppendEntriesRequest(a.Data.Term)
			slog.Info("handled AppendEntries request, sending reply to channel", "from", a.Data.ServerId)
			a.Reply <- &rpc.AppendEntriesResponse{
				ServerId:   st.GetServerIdString(),
				Term:       st.GetCurrentTerm(),
				Success:    true,
				MatchIndex: 1,
			}
			slog.Info("handled AppendEntries request", logging.From, a.Data.ServerId, logging.MessageTerm, a.Data.Term, logging.MyTerm, st.GetCurrentTerm())
		case m := <-e.Channels.VoteReply: // response from request command
			log.Printf("%s %s %s %s %s %s\n", utils.Yellow()("MAIN: New vote from"), utils.Yellow()(m.ServerId), utils.Yellow()("for term,"), utils.Yellow()(m.Term), utils.Yellow()("with value"), utils.Yellow()(m.VoteGranted))
			st.HandleVoteResponse(m.ServerId, m.VoteGranted, m.Term)
		case t := <-st.GetVoteFireChan(): // internal command
			log.Printf("%s %s\n", utils.Red()("MAIN: expiry timer fired. time:"), t)
			if st.HandleVoteTimerFired() {
				i, t := st.GetLastLogProps()
				rc.SendVoteRequests(&rpc.RequestVoteRequest{
					Term:         st.GetCurrentTerm(),
					ServerId:     st.GetServerIdString(),
					LastLogIndex: i,
					LastLogTerm:  t,
				}, e.Channels.VoteReply, e.Channels.Err)
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
			}, e.Channels.AppendEntriesReply, e.Channels.Err)
			st.HandlePingRequestSent()
		case a := <-e.Channels.AppendEntriesReply:
			log.Println("Received append entries reply", a)
		case e := <-e.Channels.Err:
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

func info(msg string, from string, to string, mTerm string, args ...any) {

	if len(from) > 0 {
		args = append(args, logging.From)
		args = append(args, from)
	}

	if len(to) > 0 {
		args = append(args, logging.To)
		args = append(args, to)
	}

	if len(mTerm) > 0 {
		args = append(args, logging.MessageTerm)
		args = append(args, mTerm)
	}

	args = append(args, logging.Me)
	args = append(args, sc.Me.Id)
	args = append(args, logging.Url)
	args = append(args, sc.Me.BaseAddress)
	args = append(args, logging.MyTerm)
	args = append(args, st.GetCurrentTerm())

	slog.Info(msg, args...)
}
