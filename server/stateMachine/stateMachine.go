package stateMachine

import "myDb/server/cfg"

type StateMachine struct {
	Term   int
	Cfg    cfg.ServerConfig
	Leader cfg.ServerIdentity
	Role   cfg.ServerRole
}
