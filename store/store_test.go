package store

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
)

func TestStore_Hash_uniqResults(t *testing.T) {
	testDBPath := "TestStore_Hash_uniqResults.db"
	os.Remove(testDBPath)
	hashCount := 10000
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := Store{}
	storage.Init(db)

	results := make(map[string]int)

	for index := 0; index < hashCount; index++ {
		hashCode, _ := storage.Hash()
		results[string(hashCode)] = index
	}
	if len(results) != hashCount {
		t.Fatalf("(len(results) = %d) != %d", len(results), hashCount)
	}

	storage.Close()
	os.Remove(testDBPath)

}

func TestStore_Hash_paramSaving(t *testing.T) {
	testDBPath := "TestStore_Hash_paramSaving.db"

	os.Remove(testDBPath)
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := Store{}
	storage.Init(db)

	hashCode, _ := storage.Hash()
	if !reflect.DeepEqual(hashCode, []byte("1")) {
		t.Fatalf("Expected: 1 got: %s", hashCode)
	}

	storage2 := Store{}
	storage2.Init(db)

	hashCode, _ = storage2.Hash()
	if !reflect.DeepEqual(hashCode, []byte("2")) {
		t.Fatalf("Expected: 2 got: %s", hashCode)
	}

	storage.Close()
	storage2.Close()
	os.Remove(testDBPath)

}

func TestStore_Save(t *testing.T) {
	testDBPath := "TestStore_Save.db"

	os.Remove(testDBPath)
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := Store{}
	storage.Init(db)
	defer storage.Close()

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
			got, err := storage.SaveLocked(tt.fullURL)
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
	os.Remove(testDBPath)
}

// Running tool: C:\Go\bin\go.exe test -benchmem -run=^$ -bench ^Benchmark_Store_Hash$

// goos: windows
// goarch: amd64
// Benchmark_Store_Hash-12    	  500000	      2617 ns/op	     531 B/op	       7 allocs/op
// PASS
// ok  	_/d_/git/Shorter/store	33.070s
// Success: Benchmarks passed.

func Benchmark_Store_Hash(b *testing.B) {
	testDBPath := "Benchmark_Store_Hash.db"

	// os.Remove(testDBPath)
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := Store{}
	storage.Init(db)
	for index := 0; index < b.N; index++ {
		b.StartTimer()
		_, err := storage.Hash()
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
	storage.Close()
	os.Remove(testDBPath)
}

// Running tool: C:\Go\bin\go.exe test -benchmem -run=^$ -bench ^Benchmark_Store_Save$

// goos: windows
// goarch: amd64
// Benchmark_Store_Save-12    	   10000	    122218 ns/op	    5058 B/op	      30 allocs/op
// PASS
// ok  	_/d_/git/Shorter/store	2.095s
// Success: Benchmarks passed.

func Benchmark_Store_Save(b *testing.B) {
	testDBPath := "Benchmark_Store_Save.db"

	os.Remove(testDBPath)
	db, _ := bolt.Open(testDBPath, 0600, nil)
	storage := Store{}
	storage.Init(db)
	b.ResetTimer()

	for index := 0; index < b.N; index++ {
		b.StopTimer()
		full := []byte(fmt.Sprintf("%d", index))
		b.StartTimer()
		_, err := storage.Save(full)
		if err != nil {
			b.Fatal(err)
		}
	}
	storage.Close()
	os.Remove(testDBPath)
}
