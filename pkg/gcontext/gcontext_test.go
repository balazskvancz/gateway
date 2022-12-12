package gcontext

import "testing"

func TestNext(t *testing.T) {
	tt := []struct {
		name     string
		fun      func(ctx *GContext)
		isCalled bool
	}{
		{
			name:     "not calling next",
			fun:      func(ctx *GContext) {},
			isCalled: false,
		},
		{
			name: "calling next",
			fun: func(ctx *GContext) {
				ctx.Next()
			},
			isCalled: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := New(nil, nil)
			tc.fun(ctx)

			if ctx.IsNextCalled() != tc.isCalled {
				t.Errorf("expected nextCalled to be: %v; got: %v\n", tc.isCalled, ctx.nextCalled)
			}
		})

	}
}

func TestNextIteration(t *testing.T) {
	tt := []struct {
		name string
		fun  func(ctx *GContext)

		expectedIndex int
	}{
		{
			name:          "the nexIteration is not called",
			fun:           func(ctx *GContext) {},
			expectedIndex: 0,
		},
		{
			name: "the nexIteration is called",
			fun: func(ctx *GContext) {
				ctx.NextIteration()
			},
			expectedIndex: 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := New(nil, nil)

			tc.fun(ctx)

			if ctx.IsNextCalled() != false {
				t.Errorf("expected false; got true")
			}

			if ctx.GetIndex() != uint8(tc.expectedIndex) {
				t.Errorf("expectedIndex: %d; got index: %d\n", tc.expectedIndex, ctx.GetIndex())
			}
		})
	}
}
