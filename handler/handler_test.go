package handler

import (
	"testing"

	"github.com/valyala/fasthttp"
)

type RequestCtxMock struct {
	path   []byte
	method []byte
}

func (mock *RequestCtx_mock) Path() []byte {
	return mock.path
}

func (mock *RequestCtx_mock) Method() []byte {
	return mock.method
}

func Benchmark_Handler_Router(t *testing.T) {
	type args struct {
		ctx *fasthttp.RequestCtx
	}
	tests := []struct {
		name string
		sh   *Handler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.sh.Router(tt.args.ctx)
		})
	}
}
