package console

import (
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-console/server"
	"github.com/vela-ssoc/vela-kit/lua"
)

var xEnv vela.Environment

func start(L *lua.LState) int {
	err := server.Instance.New(L)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}

	return 0
}

func WithEnv(env vela.Environment) {
	server.Instance.Inject(env)
	env.Set("console", lua.NewFunction(start))
}
