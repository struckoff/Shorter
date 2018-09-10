package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/struckoff/Shorter/handler"
	"github.com/valyala/fasthttp"

	"github.com/struckoff/Shorter/configuration"
)

// Структура конфига

func main() {
	conf := configuration.Configuration{}
	if err := conf.Read(os.Args[1]); err != nil {
		panic(err)
	}

	db, err := bolt.Open(conf.DBPath, 0600, nil)
	if err != nil {
		panic(err)
	}

	shorter := handler.Handler{}
	shorter.Init(db)
	fmt.Printf("Server run on %s\n", conf.Address)
	fasthttp.ListenAndServe(conf.Address, shorter.Router)

	handler.storage.Close()
	db.Close()
}
