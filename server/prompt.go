package server

import (
	"fmt"
	"github.com/vela-ssoc/vela-kit/vela"
	"strings"
)

const (
	ROOT modeType = iota
	CODE
	PROC
)

var modeValues = []string{"ROOT", "CODE", "PROC"}

type modeType int

func (m modeType) String() string {
	return modeValues[int(m)]
}

type prompt struct {
	//isRoot bool

	mode modeType
	tree []string
}

func newPrompt() *prompt {
	return &prompt{
		//isRoot: true,
		mode: ROOT,
		tree: []string{"~"},
	}
}

func (p *prompt) Len() int {
	return len(p.tree)
}

func (p *prompt) Get() string {
	return p.tree[int(p.mode)]
}

func (p *prompt) Code() string {
	if p.mode >= CODE {
		return p.tree[CODE]
	}

	return ""
}

func (p *prompt) Proc() string {
	if p.mode != PROC {
		return ""
	}
	return p.tree[PROC]
}

func (p *prompt) Add(path string) {
	switch p.mode {

	case ROOT, CODE:
		p.tree = append(p.tree, path)
		p.mode++

	case PROC:
		//todo
	}
}

func (p *prompt) Back() {
	switch p.mode {
	case ROOT:
		//todo
	case CODE, PROC:
		p.tree = p.tree[:p.Len()-1]
		p.mode--
	}
}

func (p *prompt) Last() string {
	n := len(p.tree)
	if n == 1 {
		return p.tree[0]
	}

	return p.tree[n-1]
}

func (p *prompt) useFunc(line string) []string {
	switch p.mode {

	case ROOT:
		tl := xEnv.TaskList()
		v := make([]string, len(tl))
		for idx, code := range tl {
			v[idx] = code.Name
		}
		return v

	case CODE:
		var v []string
		task := xEnv.FindTask(p.Code())
		if task == nil {
			return v
		}
		task.Range(func(r *vela.Runner) {
			v = append(v, r.Name)
		})

		return v

	case PROC:
		//todo
	}

	return empty(line)
}

func (p *prompt) Debug() string {
	return fmt.Sprintf("mode: %s\ntree: %s", p.mode, strings.Join(p.tree, "/"))
}

func (p *prompt) String() string {
	switch p.mode {
	case ROOT:
		return fmt.Sprintf(format, "~")
	case CODE:
		return fmt.Sprintf(format, p.Code())
	case PROC:
		return fmt.Sprintf(format, strings.Join(p.tree[1:], "."))
	default:
		return fmt.Sprintf(format, "~")
	}
}
