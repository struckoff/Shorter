package main

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const (
	testDBPath    = "test.db"               // Путь к тестовой БД
	serverAddress = "http://localhost:8081" // Адресс сервера приложения
)

// Нагружем HTTP-Listener GET-запросами (получаем по короткой ссылке полную)

// Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_main_GET$

// goos: linux
// goarch: amd64
// Benchmark_main_GET-8   	    5000	    237160 ns/op	   17671 B/op	     145 allocs/op
// PASS
// ok  	_/home/struckoff/Projects/godir/Shoter	1.967s
// Success: Benchmarks passed.

func Benchmark_main_GET(b *testing.B) {
	for index := 0; index < b.N; index++ {
		url := fmt.Sprintf("%s/%d", serverAddress, index)
		res, err := http.Get(url)
		if err != nil {
			b.Fatal(err.Error())
		} else if (res.StatusCode != 200) && (res.StatusCode != 404) {
			b.Fatalf("Server return: %d", res.StatusCode)
		}
	}
}

// Нагружем HTTP-Listener POST-запросами (сохраняем полную ссылку и получаем короткую)

// Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_main_POST$

// goos: linux
// goarch: amd64
// Benchmark_main_POST-8   	    5000	    524506 ns/op	   18608 B/op	     155 allocs/op
// PASS
// ok  	_/home/struckoff/Projects/godir/Shoter	2.697s
// Success: Benchmarks passed.

func Benchmark_main_POST(b *testing.B) {
	for index := 0; index < b.N; index++ {
		b.StopTimer()
		msg := fmt.Sprintf("%d", index)
		d, _ := time.ParseDuration("3s")
		time.Sleep(d)
		b.StartTimer()
		res, err := http.Post(serverAddress, "text/plain", bytes.NewBuffer([]byte(msg)))
		if err != nil {
			panic(err)
		} else if res.StatusCode != 200 {
			b.Fatalf("Server return %d instead of 200", res.StatusCode)
		}
	}
}
