package server

import (
	"context"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/vela-ssoc/vela-kit/vela"
	"github.com/vela-ssoc/vela-kit/lua"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

type session struct {
	auth   bool
	pass   string
	err    error
	parent context.Context
	ctx    context.Context
	stop   context.CancelFunc
	cfg    readline.Config
	conn   net.Conn
	pmt    *prompt
	liner  *readline.Instance
}

func (sess *session) init() {

	cfg := readline.Config{
		Prompt:              sess.pmt.String(),
		HistoryFile:         os.TempDir() + "/.vela-console.tmp",
		InterruptPrompt:     "^cfg",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		AutoComplete:        newCompleter(sess.pmt),
		FuncFilterInputRune: sess.filterInput,
	}
	liner, err := readline.HandleConn(cfg, sess.conn)

	sess.ctx, sess.stop = context.WithCancel(context.Background())

	sess.auth = false
	sess.pass = "rock.good"
	sess.cfg = cfg
	sess.err = err
	sess.liner = liner
}

func (sess *session) Close() {
	sess.conn.Close()
}

func (sess *session) Println(v string) {
	fmt.Fprintln(sess.liner.Stdout(), v)
	//sess.liner.Stdout().Write(lua.S2B(v))
}

func (sess *session) Printf(format string, v ...interface{}) {
	sess.Println(fmt.Sprintf(format, v...))
}

func (sess *session) Invalid(format string, v ...interface{}) {
	f := time.Now().Format(time.RFC822) + " " + format
	sess.Printf(f, v...)
}

func (sess *session) Auth(line string) bool {

	//已经登陆
	if xEnv.IsDebug() || sess.auth {
		return true
	}

	//判断是否为登陆请求
	if !strings.HasPrefix(line, "auth ") {
		sess.Println("no login")
		goto exit
	}

	//比较密码
	if line[5:] == sess.pass {
		sess.auth = true
		sess.Println("login succeed")
		goto exit
	}

	//错误密码
	sess.Invalid("invalid pass")

exit:
	return false
}

func (sess *session) AutoPrompt() {
	sess.liner.SetPrompt(sess.pmt.String())
}

func (sess *session) filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func (sess *session) Use(line string) {
	line = strings.TrimSpace(line)

	if line == "" {
		sess.Invalid("invalid command options , usage: use service")
		return
	}

	switch sess.pmt.mode {
	case ROOT:
		if xEnv.FindTask(line) == nil {
			sess.Invalid("not found %s code", line)
			return
		}
		goto done

	case CODE:
		if code := xEnv.FindTask(sess.pmt.Code()); code != nil && code.Exist(line) {
			goto done
		}
		sess.Invalid("not found %s proc: %s", sess.pmt.Code(), line)
		return

	case PROC:
		sess.Invalid("can't use in proc mode")
		return
	}

done:
	sess.pmt.Add(line)
	sess.liner.SetPrompt(sess.pmt.String())
}

func (sess *session) Load(line string, env vela.Environment) {
	path := strings.TrimSpace(line)
	if path == "" {
		sess.Invalid("not found")
		return
	}

	chunk, err := ioutil.ReadFile(fmt.Sprintf("%s/%s",
		Instance.cfg.Script, path))

	if err != nil {
		sess.Invalid(err.Error())
		return
	}

	//取后缀
	name := lua.FileSuffix(path)

	err = xEnv.LoadTask(name, chunk, sess)
	if err != nil {
		sess.Invalid("%s", err)
		return
	}
	sess.Printf("load %s succeed", path)
}

func (sess *session) Pwd() {
	sess.Printf("%s", sess.pmt.Debug())
	sess.Println(sess.pmt.String())
}

func (sess *session) Usage() {
	sess.Println("use： 切换到当前服务配置选项")
	sess.Println("list: 查看当前配置服务列表")
	sess.Println("load: 加载外部脚本")
	sess.Println("help: 帮助信息")

	if sess.pmt.mode == PROC {
		sess.Println("show: 显示配置参数")
		sess.Println(".:  点开头的说明调用服务帮顶的lua方法 如: .debug(123)")
		code := xEnv.FindCode(sess.pmt.Code())
		if code == nil {
			sess.Invalid("not found %s task", sess.pmt.Code())
			return
		}

		proc := code.Get(sess.pmt.Proc())
		if proc == nil {
			sess.Invalid("not found %s.%s error", sess.pmt.Code(), sess.pmt.Proc())
			return
		}
		proc.Data.Help(sess)
	}
}

func (sess *session) Clear() {
	sess.liner.Clean()
}

func (sess *session) List() {
	switch sess.pmt.mode {
	case ROOT:
		tl := xEnv.TaskList()
		for _, task := range tl {
			sess.Printf("+----------------------------------------------------+")
			sess.Printf("|配置: %-46s|", task.Name)
			sess.Printf("+----------------------------------------------------+")
			sess.Printf("|哈希: %-46s|", task.Hash)
			sess.Printf("+----------------------------------------------------+")
			sess.Printf("|状态: %-46s|", task.Status)
			sess.Printf("+----------------------------------------------------+")
			sess.Printf("|外链: %-46s|", task.Link)
			sess.Printf("+----------------------------------------------------+")
			sess.Println(display("|", task))
		}

	case CODE:
		task := xEnv.FindTask(sess.pmt.Code())
		if task == nil {
			sess.Invalid("not found %s", sess.pmt.tree[1])
			return
		}
		sess.Printf("+----------------------------------------------------+")
		sess.Println(display("|", task))

	case PROC:
		//todo

	}
}

func (sess *session) Delete(line string) {
	line = strings.TrimSpace(line)
	switch sess.pmt.mode {

	case ROOT:
		xEnv.RemoveTask(line, vela.CONSOLE)

	case CODE:
		//code := service.GetCodeVM(sess.pmt.Code())
		//if code != nil {
		//	code.Del(line)
		//}
		sess.Invalid("can't use del in code mode")
		return

	case PROC:
		sess.Invalid("can't use del in proc mode")
		return
	}
}

func (sess *session) Do(env vela.Environment) {
	defer sess.AutoPrompt()

	sess.Println("请输入要执行的lua代码:")

	sess.liner.SetPrompt("")
	chunk := make([]string, 0)
	sess.liner.SetVimMode(true)

	for {
		line, err := sess.liner.Readline()
		if err != nil {
			return
		}

		if line == "exit" {
			goto exit
		}

		chunk = append(chunk, line)
	}

exit:
	co := env.Coroutine()
	err := env.DoString(co, strings.Join(chunk, "\n"))

	if err != nil {
		sess.Printf("%v", err)
	} else {
		sess.Println("succeed")
	}
	env.Free(co)
}

func (sess *session) DoString(line string, env vela.Environment) {
	co := env.Coroutine()
	err := env.DoString(co, line)
	if err != nil {
		sess.Invalid("%v", err)
	} else {
		sess.Println("succeed")
	}
	env.Free(co)
}

func (sess *session) Quit() {
	switch sess.pmt.mode {
	case ROOT:
		sess.stop()
	case CODE, PROC:
		sess.pmt.Back()
		sess.liner.SetPrompt(sess.pmt.String())
	}
}

func (sess *session) doInterpreter(line string) []byte {

	//代码生成
	var chunk string
	switch line {
	case "start":
		chunk = fmt.Sprintf(`proc.start(proc.get("%s"))`, sess.pmt.Proc())
	case "close":
		chunk = fmt.Sprintf(`proc.close(proc.get("%s"))`, sess.pmt.Proc())
	default:
		chunk = fmt.Sprintf(`proc.get("%s").%s`, sess.pmt.Proc(), line)
	}

	return lua.S2B(chunk)
}

func (sess *session) doFunc(line string) {
	line = strings.TrimSpace(line)

	switch sess.pmt.mode {
	case ROOT, CODE:
		//todo

	case PROC:

		//历史执行代码
		//co := service.GetCodeVM(sess.pmt.Code())
		//if co == nil {
		//	sess.Invalid("not found %s.lua", sess.pmt.Code())
		//	return
		//}

		////执行代码
		//chunk := sess.doInterpreter(line)
		//err := code.DoProcFunc(chunk, xcall.Rock, sess)
		//if err != nil {
		//	sess.Invalid("%v", err)
		//	return
		//}
		//sess.Invalid("succeed")
	}

}

func (sess *session) Show() {
	switch sess.pmt.mode {

	case ROOT, CODE:
		sess.List()

	case PROC:
		proc, err := xEnv.FindProc(sess.pmt.Code(), sess.pmt.Proc())
		if err != nil {
			sess.Invalid("code:%s proc: %s err: %v", sess.pmt.Code(), sess.pmt.Get(), err)
			return
		}

		proc.Data.Show(sess)
	}

}

func (sess *session) handler(env vela.Environment) {
	defer sess.Close()

	if sess.err != nil {
		return
	}

	for {
		select {

		case <-sess.ctx.Done():
			return
		case <-sess.parent.Done():
			return

		default:

			line, e := sess.liner.Readline()
			if e == readline.ErrInterrupt {
				if len(line) == 0 {
					return
				}
				continue
			}

			if e == io.EOF {
				return
			}

			if !sess.Auth(line) {
				continue
			}

			line = strings.TrimSpace(line)
			switch {
			case line == "":
			case line == "?":
				sess.Usage()
			case line == "help":
				sess.Usage()
			case line == "quit":
				sess.Quit()

			case line == "list":
				sess.List()
			case line == "ls":
				sess.List()

			case line == "pwd":
				sess.Pwd()

			case line == "clear":
				sess.Clear()
			case line == "show":
				sess.Show()
			case line == "do":
				sess.Do(env)

			case strings.HasPrefix(line, "load "):
				sess.Load(line[5:], env)
			case strings.HasPrefix(line, "use "):
				sess.Use(line[4:])
			case strings.HasPrefix(line, "del "):
				sess.Delete(line[4:])
			case strings.HasPrefix(line, "do "):
				sess.DoString(line[3:], env)

			case strings.HasPrefix(line, ".") && sess.pmt.mode == PROC:
				sess.doFunc(line[1:])

			default:
				sess.Invalid("Invalid command %s", line)
			}
		}
	}
}
