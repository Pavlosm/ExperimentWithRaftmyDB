package main

import (
	"fmt"
	"myDb/server/cfg"
	"myDb/server/stateMachine"
	"os"
	"strconv"
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
}
