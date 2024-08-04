package cfg

import "strconv"

type ServerIdentity struct {
	Id          string
	Port        int
	BaseAddress string
}

func (s ServerIdentity) GetUrl() string {
	return s.BaseAddress + ":" + strconv.Itoa(s.Port)
}
