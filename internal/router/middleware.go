package router

import (
	"github.com/balazskvancz/gateway/internal/gcontext"
)

type middlewareChain struct {
	chain *[]HandlerFunc
}

type Middleware struct {
	part    string
	handler HandlerFunc
}

// Creates a new handleChain.
func createNewMWChain(handler HandlerFunc, mw ...HandlerFunc) *middlewareChain {
	if handler == nil {
		return nil
	}

	chain := []HandlerFunc{}

	// Adding the middlwares to the slice.
	for _, mwH := range mw {
		chain = append(chain, mwH)
	}

	// Appends the handler to the end.
	chain = append(chain, handler)

	return &middlewareChain{
		chain: &chain,
	}
}

// Gets the last element of the MW chain, which is the handler itself.
func (mw *middlewareChain) getLast() HandlerFunc {
	return (*mw.chain)[len(*mw.chain)-1]
}

// Executes the chain of middlewares. The last element is the handler itself.
func (mw *middlewareChain) run(ctx *gcontext.GContext) {
	canContinue := true

	for _, handler := range *mw.chain {
		if ctx.GetIndex() == 0 || canContinue {
			handler(ctx)

			canContinue = ctx.IsNextCalled()
			ctx.NextIteration()
			continue
		}

		break
	}
}
