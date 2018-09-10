package handler

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/struckoff/Shorter/store"
	"github.com/valyala/fasthttp"
)

// Handler - keeps application logic
type Handler struct {
	storage store.Store
}

// Init initialize storage
func (sh *Handler) Init(db *bolt.DB) error {
	sh.storage = store.Store{}
	err := sh.storage.Init(db)
	return err
}

// doPost - handle POST requests. Checks if short url already exists, if its not - generate a new one
func (sh *Handler) doPost(ctx *fasthttp.RequestCtx) {
	fullURL := ctx.PostBody()
	if len(fullURL) == 0 {
		ctx.Error("Body is empty", fasthttp.StatusBadRequest)
	}
	short, err := sh.storage.Save(fullURL)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	fmt.Fprintf(ctx, "Short url: %s/%s", ctx.Host(), short)
	return
}

// doGet - handle GET requests. Checks if short url exists returns it by pure text, if its not - returns 404
func (sh *Handler) doGet(ctx *fasthttp.RequestCtx) {
	short := ctx.Path()[1:]
	if full, err := sh.storage.FullURL(short); full != nil {
		// ctx.Redirect(full, fasthttp.StatusMovedPermanently)
		fmt.Fprintf(ctx, "Full url: %s", full)
	} else if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else {
		ctx.NotFound()
	}
	return
}

func (sh *Handler) doDefault(ctx *fasthttp.RequestCtx) {
	ctx.Error("Method not allowed!", fasthttp.StatusMethodNotAllowed)
}

// Router - routes requests by HTTP-method
func (sh *Handler) Router(ctx *fasthttp.RequestCtx) {
	if ctx.IsPost() {
		sh.doPost(ctx)
		return
	}
	if ctx.IsGet() {
		sh.doGet(ctx)
		return
	}
	sh.doDefault(ctx)
	return
}

// Close - finish working with storage properly
func (sh *Handler) Close() {
	sh.storage.Close()
}
