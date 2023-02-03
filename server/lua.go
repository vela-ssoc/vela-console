package server

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/lua"
)

var xEnv vela.Environment

func (s *Serv) Inject(env vela.Environment) {
	xEnv = env

	Instance = newServ()
	xEnv.Register(Instance)

	env.Set("print", lua.NewFunction(s.output))
}
