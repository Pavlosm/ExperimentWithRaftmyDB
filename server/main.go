package main

import (
	"flag"
	"fmt"
	"log/slog"
	"myDb/rpc"
	"myDb/server/cfg"
	"myDb/server/environment"
	"myDb/server/logging"
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

	st = state.NewDefaultStateMachine(sc)

	s = *rpcServer.NewRpcServer(e)

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
		case v := <-e.Channels.VoteCmd:
			onReceivedExternalVoteRequest(v)
		case a := <-e.Channels.AppendEntriesCmd:
			onReceivedExternalAppendEntriesRequest(a)
		case m := <-e.Channels.VoteReply: // response from request command
			onReceivedExternalVoteReply(m)
		case <-st.GetVoteFireChan():
			onInternalVoteTimerFired()
		case <-st.GetPingFireChan():
			onInternalPingTimerFired()
		case a := <-e.Channels.AppendEntriesReply:
			onReceivedExternalAppendEntriesReply(a)
		case c := <-e.Channels.CommandCmd:
			onReceivedExternalCommandRequest(c)
		case e := <-e.Channels.Err:
			onError(e)
		}
	}
}

func onReceivedExternalVoteRequest(v utils.WithReplyChan[*rpc.RequestVoteRequest, *rpc.RequestVoteResponse]) {
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

	info("sent vote response", "", v.Data.ServerId, "", logging.VoteGranted, granted)
}

func onReceivedExternalAppendEntriesRequest(a utils.WithReplyChan[*rpc.AppendEntriesRequest, *rpc.AppendEntriesResponse]) {

	var logs []state.Log

	for _, e := range a.Data.Entries {
		logs = append(logs, state.Log{
			Term:    e.Term,
			Index:   e.Index,
			Command: e.Command,
		})
	}

	st.HandleAppendEntriesRequest(a.Data.Term, a.Data.ServerId, logs)

	a.Reply <- &rpc.AppendEntriesResponse{
		ServerId:   st.GetServerIdString(),
		Term:       st.GetCurrentTerm(),
		Success:    true,
		MatchIndex: 1,
	}

	info("handled AppendEntries request", a.Data.ServerId, "", string(a.Data.Term))
}

func onReceivedExternalVoteReply(m *rpc.RequestVoteResponse) {
	info("new vote received", m.ServerId, "", string(m.Term), logging.VoteGranted, m.VoteGranted)
	st.HandleVoteResponse(m.ServerId, m.VoteGranted, m.Term)
}

func onInternalVoteTimerFired() {

	if !st.HandleVoteTimerFired() {
		return
	}

	infoShort("becoming CANDIDATE Sending vote request to all servers")

	i, t := st.GetLastLogProps()

	rc.SendVoteRequests(&rpc.RequestVoteRequest{
		Term:         st.GetCurrentTerm(),
		ServerId:     st.GetServerIdString(),
		LastLogIndex: i,
		LastLogTerm:  t,
	}, e.Channels.VoteReply, e.Channels.Err)
}

func onInternalPingTimerFired() {
	infoShort("Sending append entries request to all servers")

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
}

func onReceivedExternalAppendEntriesReply(a *rpc.AppendEntriesResponse) {
	info("Received append entries reply", a.ServerId, "", strconv.FormatInt(a.Term, 10))
}

func onReceivedExternalCommandRequest(c utils.WithReplyChan[*rpc.CommandRequest, *rpc.CommandResponse]) {

	if st.GetServerRole() != state.Leader {
		c.Reply <- &rpc.CommandResponse{
			Success:  false,
			LeaderId: st.GetLeaderId(),
		}
		infoShort("not a leader, sending error response")
		return
	}

	l, err := st.HandleCommandRequest(c.Data.Command)

	if err != nil {
		c.Reply <- &rpc.CommandResponse{
			Success:  false,
			LeaderId: st.GetLeaderId(),
		}
		return
	}

	i, t := st.GetLastLogProps()

	rc.SendAppendEntries(&rpc.AppendEntriesRequest{
		Term:         st.GetCurrentTerm(),
		ServerId:     st.GetServerIdString(),
		PrevLogIndex: i,
		PrevLogTerm:  t,
		LeaderCommit: 0,
		Entries: []*rpc.Entry{
			{
				Term:    t,
				Index:   i,
				Command: l.Command,
			},
		},
	}, e.Channels.AppendEntriesReply, e.Channels.Err)

	c.Reply <- &rpc.CommandResponse{
		Success:  true,
		LeaderId: st.GetLeaderId(),
	}

	infoShort("handled Command request")
}

func onError(e error) {
	slog.Error("An error ocurred", "error", e)
}

func startRpcServer() {
	flag.Parse()

	p := int32(sc.Me.Port)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		panic(err)
	}
	gs := grpc.NewServer()

	rpc.RegisterRaftServiceServer(gs, &s)
	msg := "MAIN: server listening at " + lis.Addr().String()
	slog.Info(msg)

	if err := gs.Serve(lis); err != nil {
		slog.Error("failed to serve", "error", err)
		panic(err)
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

func infoShort(msg string, args ...any) {
	slog.Info(msg, args...)
}
