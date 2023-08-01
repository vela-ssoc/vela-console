package server

import (
	"bytes"
	"context"
	"github.com/vela-ssoc/vela-kit/audit"
	"github.com/vela-ssoc/vela-kit/lua"
	"net"
	"os"
)

type Serv struct {
	cfg    *config
	ctx    context.Context
	cancel context.CancelFunc
	//ln     *ipc.Listener
	ln net.Listener
}

func newServ() *Serv {
	cfg := normal()
	ctx, cancel := context.WithCancel(context.Background())
	return &Serv{cfg: cfg, ctx: ctx, cancel: cancel}
}

func (s *Serv) Name() string {
	return "console.server"
}

func (s *Serv) NewSession(conn net.Conn) *session {

	//会话
	sess := &session{conn: conn, pmt: newPrompt(), parent: s.ctx}

	//初始化
	sess.init()

	return sess
}

func (s *Serv) Accept() {
	for {
		select {
		case <-s.ctx.Done():
			xEnv.Error("exit vela-console ...")
			return
		default:
			conn, e := s.ln.Accept()
			if e != nil {
				xEnv.Errorf("vela-console accept fail , %v", e)
				return
			}
			xEnv.Spawn(0, func() {
				s.NewSession(conn).handler(xEnv)
			})
		}
	}
}

func (s *Serv) remove() {
	if s.cfg.Network == "unix" {
		os.Remove(s.cfg.Address)
	}

	s.ln = nil
}

func (s *Serv) Do() error {
	if s.ln != nil {
		goto accept
	}

	if e := s.Listen(); e != nil {
		return e
	}

accept:
	s.ctx, s.cancel = context.WithCancel(context.Background())
	xEnv.Spawn(0, s.Accept)
	return nil
}

func (s *Serv) Close() error {
	defer s.remove()
	s.cancel()

	if s.ln == nil {
		return nil
	}

	return s.ln.Close()
}

func (s *Serv) Listen() error {
	if s.cfg.Network == "unix" {
		s.remove()
	}

	ln, err := net.Listen(s.cfg.Network, s.cfg.Address)
	if err != nil {
		return err
	}

	s.ln = ln
	return nil
}

func (s *Serv) New(L *lua.LState) error {
	L.CheckTable(1).Range(func(key string, val lua.LValue) {
		switch key {
		case "enable":
			s.cfg.Enable = lua.CheckBool(L, val)
		case "network":
			s.cfg.Network = val.String()
		case "address":
			s.cfg.Address = val.String()
		case "script":
			s.cfg.Script = val.String()
		}
	})

	//关闭
	if !s.cfg.Enable {
		return s.Close()
	}

	//新建
	if s.ln == nil {
		return s.Do()
	}

	//重启
	if s.ln != nil {

		addr := s.ln.Addr()
		if addr.Network() == s.cfg.Network && addr.String() == s.cfg.Address {
			s.cancel()
			return s.Do()
		}

		ev := audit.NewEvent("vela-console").Subject("console重启").From("startup")
		if err := s.Close(); err != nil {
			ev.Msg("vela-console 关闭失败").E(err).Log().Put()
		} else {
			ev.Msg("vela-console 关闭成功").Log().Put()
		}

	}
	return s.Do()
}

func (s *Serv) output(L *lua.LState) int {
	sess, ok := L.Metadata(0).(*session)
	if !ok {
		return 0
	}
	n := L.GetTop()
	if n == 0 {
		return 0
	}

	var buff bytes.Buffer

	for i := 1; i <= n; i++ {
		buff.WriteString(L.Get(i).String())
	}

	sess.Println(buff.String())
	return 0
}
