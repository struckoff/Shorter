package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/valyala/fasthttp"

	"github.com/struckoff/Shorter/configuration"
	"github.com/struckoff/Shorter/handler"
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
	defer shorter.Close()

	fmt.Printf("Server run on %s\n", conf.Address)
	fasthttp.ListenAndServe(conf.Address, shorter.Router)
}
