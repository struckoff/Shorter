package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
)

///////////////////////////////////////////////////////////
// Host:

// OS: Arch Linux
// Kernel: x86_64 Linux 4.18.5-arch1-1-ARCH
// CPU: Intel Core i7-7700K @ 8x 4.5GHz
// RAM: 7325MiB / 24026MiB
// SSD
///////////////////////////////////////////////////////////

const (
	testDB_path   = "test.db"               // Путь к тестовой БД
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
		msg := fmt.Sprintf("%d", index)
		res, err := http.Post(serverAddress, "text/plain", bytes.NewBuffer([]byte(msg)))
		if err != nil {
			panic(err)
		} else if res.StatusCode != 200 {
			b.Fatalf("Server return %d instead of 200", res.StatusCode)
		}
	}
}

// Бенчмарк функции получения хэша для сборки корооткой ссылки

// Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Store_Hash$

// goos: linux
// goarch: amd64
// Benchmark_Store_Hash-8   	  300000	     16464 ns/op	    8757 B/op	       9 allocs/op
// PASS
// ok  	_/home/struckoff/Projects/godir/Shoter	5.294s
// Success: Benchmarks passed.

func Benchmark_Store_Hash(b *testing.B) {
	defer os.Remove(testDB_path)
	db, _ := bolt.Open(testDB_path, 0600, nil)
	store := Store{}
	store.Init(db)
	b.ResetTimer()

	b.StartTimer()
	for index := 0; index < b.N; index++ {
		_, err := store.Hash()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

// Бенчмарк функции сохранения полной ссылки и получения короткой
// Полная ссылка сохраняется в БД, для нее генерируется и возвращается хэш

// Running tool: /usr/bin/go test -benchmem -run=^$ -bench ^Benchmark_Store_Save$

// goos: linux
// goarch: amd64
// Benchmark_Store_Save-8   	  200000	     17582 ns/op	    7227 B/op	      19 allocs/op
// PASS
// ok  	_/home/struckoff/Projects/godir/Shoter	19.250s
// Success: Benchmarks passed.

func Benchmark_Store_Save(b *testing.B) {
	defer os.Remove(testDB_path)
	db, _ := bolt.Open(testDB_path, 0600, nil)
	store := Store{}
	store.Init(db)
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		b.StopTimer()
		full := []byte(fmt.Sprintf("%d", index))
		b.StartTimer()
		_, err := store.Save(full)
		if err != nil {
			b.Fatal(err)
		}
	}
	store.Close()
}

func TestStore_Save(t *testing.T) {
	defer os.Remove(testDB_path)
	db, _ := bolt.Open(testDB_path, 0600, nil)
	store := Store{}
	store.Init(db)
	defer store.Close()

	type args struct {
		fullURL []byte
	}
	tests := []struct {
		fullURL []byte
		expect  []byte
	}{
		{[]byte("http://tt.t"), []byte("1")},
		{[]byte("http://tt.t/ttt"), []byte("2")},
		{[]byte("http://tt.t/ddd"), []byte("3")},
		{[]byte("http://tt.t"), []byte("1")},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(string(tt.fullURL), func(t *testing.T) {
			got, err := store.SaveLocked(tt.fullURL)
			if err != nil {
				t.Errorf("Store.Save() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("Store.Save() = %v, want %v", string(got), string(tt.expect))
			}
		})
	}
	db.Close()
}
