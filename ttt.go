package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/struckoff/Shorter/store"
)

const (
	testDBPath    = "test.db"               // Путь к тестовой БД
	serverAddress = "http://localhost:8081" // Адресс сервера приложения
)

func main() {
	defer os.Remove(testDBPath)
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := store.Store{}
	storage.Init(db)
	defer storage.Close()

	for i := 0; i < 1000; i++ {
		res, _ := storage.Hash()
		fmt.Println(string(res))
	}
}
