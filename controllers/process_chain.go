package controllers

import "sync"

type (
	ProcessContext struct {
		Ctx           interface{}
		Param         map[string]interface{}
		PostProcessFn func(*ProcessContext, []byte, error)
	}

	ProcessUnit  func(ctx *ProcessContext) ([]byte, error)
	ProcessChain []ProcessUnit
)

func NewProcessChain() ProcessChain {
	return make(ProcessChain, 0)
}

func (pc ProcessChain) Add(unit ProcessUnit) ProcessChain {
	return append(pc, unit)
}

func (pc ProcessChain) Process(ctx interface{}, postProcessFn func(*ProcessContext, []byte, error)) {
	var (
		retBuf []byte
		err    error
		pcCtx  = GetProcessContext().SetCtx(ctx)
	)

	for _, fn := range pc {
		retBuf, err = fn(pcCtx)
		if len(retBuf) != 0 || err != nil {
			break
		}
	}

	postProcessFn(pcCtx, retBuf, err)
	PutProcessContext(pcCtx)
	return
}

func NewProcessContext() *ProcessContext {
	return &ProcessContext{
		Ctx:   nil,
		Param: make(map[string]interface{}),
	}
}

func (pcCtx *ProcessContext) SetCtx(ctx interface{}) *ProcessContext {
	pcCtx.Ctx = ctx
	return pcCtx
}

var processContextPool sync.Pool

func init() {
	processContextPool = sync.Pool{New: func() interface{} {
		return NewProcessContext()
	}}
}

func GetProcessContext() *ProcessContext {
	return processContextPool.Get().(*ProcessContext)
}

func PutProcessContext(ctx *ProcessContext) {
	processContextPool.Put(ctx)
}
