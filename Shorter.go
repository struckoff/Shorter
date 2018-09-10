package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/valyala/fasthttp"
)

// Структура конфига

func main() {
	conf := Configuration{}
	if err := conf.Read(os.Args[1]); err != nil {
		panic(err)
	}

	db, err := bolt.Open(conf.DBPath, 0600, nil)
	if err != nil {
		panic(err)
	}

	handler := Handler{}
	handler.Init(db)
	fmt.Printf("Server run on %s\n", conf.Address)
	fasthttp.ListenAndServe(conf.Address, handler.Router)

	handler.store.Close()
	db.Close()
}
