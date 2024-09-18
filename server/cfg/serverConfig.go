package cfg

type ServerConfig struct {
	Me          ServerIdentity
	Servers     []ServerIdentity
	QuorumCount int
}

func NewServerConfig(myId NodeId, m map[NodeId]ServerIdentity) ServerConfig {

	sc := ServerConfig{
		Servers: []ServerIdentity{},
	}

	for k, si := range m {
		if k == myId {
			sc.Me = si
		} else {
			sc.Servers = append(sc.Servers, si)
		}
	}

	sc.QuorumCount = len(m) / 2

	return sc
}

func (s *ServerConfig) MajorityNum() int {
	return (len(s.Servers) / 2) + 1
}
