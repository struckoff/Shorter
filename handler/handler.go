package handler

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/struckoff/Shorter/store"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	storage store.Store
}

func (sh *Handler) Init(db *bolt.DB) error {
	sh.storage = store.Store{}
	err := sh.storage.Init(db)
	return err
}

// Обработка POST-запросов
// Если ссылка уже есть в БД - возвращется оттуда, иначе генерируется
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

// Обработка GET-запросов
// Если ссылка уже есть в БД - возвращется текстом, иначе возвращается 404
func (sh *Handler) doGet(ctx *fasthttp.RequestCtx) {
	short := ctx.Path()[1:]
	if full, err := sh.storage.GetFull(short); full != nil {
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

// Рутер HTTP-методов
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
