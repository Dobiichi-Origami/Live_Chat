package controllers

type (
	ProcessContext struct {
		Ctx           interface{}
		Param         map[string]interface{}
		PostProcessFn func(*ProcessContext, []byte, error)
	}

	ProcessUnit  func(ctx *ProcessContext) ([]byte, error)
	ProcessChain []ProcessUnit
)

func NewProcessChain(size int) ProcessChain {
	return make(ProcessChain, size, size)
}

func (pc ProcessChain) Add(unit ProcessUnit) ProcessChain {
	return append(pc, unit)
}

func (pc ProcessChain) Process(ctx *ProcessContext) {
	var (
		retBuf []byte
		err    error
	)

	for _, fn := range pc {
		retBuf, err = fn(ctx)
		if len(retBuf) != 0 || err != nil {
			break
		}
	}

	ctx.PostProcessFn(ctx, retBuf, err)
	return
}
