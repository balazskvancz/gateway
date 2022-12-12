package router

import (
	"testing"

	"github.com/balazskvancz/gateway/pkg/gcontext"
)

func TestMiddleware(t *testing.T) {
	firstChain := createNewMWChain(
		func(g *gcontext.GContext) {}, // the handler
		func(g *gcontext.GContext) {},
		func(g *gcontext.GContext) {},
		func(g *gcontext.GContext) {},
	)

	secondChain := createNewMWChain(
		func(g *gcontext.GContext) {}, // the handler
		func(g *gcontext.GContext) { g.Next() },
		func(g *gcontext.GContext) {},
		func(g *gcontext.GContext) {},
	)

	thirdChain := createNewMWChain(
		func(g *gcontext.GContext) {}, // the handler
		func(g *gcontext.GContext) { g.Next() },
		func(g *gcontext.GContext) { g.Next() },
		func(g *gcontext.GContext) {},
	)

	fourthChain := createNewMWChain(
		func(g *gcontext.GContext) {}, // the handler
		func(g *gcontext.GContext) { g.Next() },
		func(g *gcontext.GContext) { g.Next() },
		func(g *gcontext.GContext) { g.Next() },
	)

	tt := []struct {
		name          string
		chain         *middlewareChain
		expectedIndex uint8
	}{
		{
			name:          "the chain stops after the first mw",
			chain:         firstChain,
			expectedIndex: 1,
		},
		{
			name:          "the chain stops after the the second mw",
			chain:         secondChain,
			expectedIndex: 2,
		},
		{
			name:          "the chain stops after the the third mw",
			chain:         thirdChain,
			expectedIndex: 3,
		},
		{
			name:          "the chain passes all the mws, and calls the handler",
			chain:         fourthChain,
			expectedIndex: 4,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := gcontext.New(nil, nil)

			tc.chain.run(ctx)

			if ctx.GetIndex() != tc.expectedIndex {
				t.Errorf("expected index: %d, got: %d\n", tc.expectedIndex, ctx.GetIndex())
			}

		})

	}
}
