package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"github.com/valyala/fasthttp"
)

// Алфавит
const ALPH = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Адрес конечного символа алфавита
const ALPH_LAST = int64(len(ALPH) - 1)

// Структура конфига
type Configuration struct {
	DBPath  string `toml:"DBPath"`
	Address string `toml:"Address"`
}

// Структура хранилища
type Store struct {
	lastID    int64      // Последний ID для генерации хэша
	idChannel chan int64 // который отдает ID для гарантированой инкрментации в одной точке
	db        *bolt.DB
	wg        sync.WaitGroup
}

func (s *Store) Init(db *bolt.DB) error {
	s.idChannel = make(chan int64)
	s.db = db
	s.wg = sync.WaitGroup{}
	err := s.db.Batch(func(tx *bolt.Tx) error {
		var err error
		// Маппинг короттких к полным ссылками
		_, err = tx.CreateBucketIfNotExists([]byte("shortToFull"))
		if err != nil {
			panic(err)
		}
		// Маппинг полных к коротким ссылками
		_, err = tx.CreateBucketIfNotExists([]byte("fullToShort"))
		if err != nil {
			panic(err)
		}

		// Корзина с состояниями
		props := tx.Bucket([]byte("Props"))
		if props != nil {
			LastID := props.Get([]byte("LastID"))
			s.lastID = int64(binary.LittleEndian.Uint64(LastID))
		}

		return err
	})
	if err != nil {
		return err
	}
	// Счетчик для инкрементирования ID по запросу
	go func(idChannel chan int64) {
		for {

			s.lastID++
			idChannel <- s.lastID

			go s.db.Update(func(tx *bolt.Tx) error {
				bucket, err := tx.CreateBucketIfNotExists([]byte("Props"))
				if err != nil {
					panic(err)
				}
				b := make([]byte, 8)
				binary.LittleEndian.PutUint64(b, uint64(s.lastID))
				err = bucket.Put([]byte("LastID"), b)

				return err
			})
		}
	}(s.idChannel)

	return nil
}

func (s *Store) Close() {
	s.wg.Wait()
	// close(s.idChannel)
}

// Сохранение полной ссылки в БД,
// получение хэша для составления короткой ссылки
func (s *Store) Save(fullURL []byte) ([]byte, error) {

	if short, err := s.getShort(fullURL); short != nil {
		return short, nil
	} else if err != nil {
		return nil, err
	}

	short, err := s.Hash()
	if err != nil {
		return nil, err
	}

	// Создаем горутину для записи ссылок в БД
	go s.db.Update(func(tx *bolt.Tx) error {
		defer s.wg.Done()
		s.wg.Add(1)
		shortToFull := tx.Bucket([]byte("shortToFull"))
		fullToShort := tx.Bucket([]byte("fullToShort"))
		if err := shortToFull.Put(short, fullURL); err != nil {
			fmt.Errorf("INSERT err: %s", err)
			return err
		}
		if err := fullToShort.Put(fullURL, short); err != nil {
			fmt.Errorf("INSERT err: %s", err)
			return err
		}

		return nil
	})

	return short, err
}

// Тоже самое что и Save(), но ожидается завершение записи в БД
func (s *Store) SaveLocked(fullURL []byte) ([]byte, error) {
	if short, err := s.getShort(fullURL); short != nil {
		return short, nil
	} else if err != nil {
		return nil, err
	}

	short, err := s.Hash()
	if err != nil {
		return nil, err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		shortToFull := tx.Bucket([]byte("shortToFull"))
		fullToShort := tx.Bucket([]byte("fullToShort"))
		if err := shortToFull.Put(short, fullURL); err != nil {
			fmt.Errorf("INSERT err: %s", err)
			return err
		}
		if err := fullToShort.Put(fullURL, short); err != nil {
			fmt.Errorf("INSERT err: %s", err)
			return err
		}
		return nil
	})

	return short, err
}

// Получение полной ссылки по короткой
func (s *Store) getFull(short []byte) ([]byte, error) {
	full := []byte{}
	err := s.db.View(func(tx *bolt.Tx) error {
		shortToFull := tx.Bucket([]byte("shortToFull"))
		full = shortToFull.Get(short)
		return nil
	})

	return full, err
}

// Получение короткой ссылки по полной
func (s *Store) getShort(full []byte) ([]byte, error) {
	short := []byte{}
	err := s.db.View(func(tx *bolt.Tx) error {
		fullToShort := tx.Bucket([]byte("fullToShort"))
		short = fullToShort.Get(full)
		return nil
	})

	return short, err
}

// Генерируем Хэш
func (s *Store) Hash() ([]byte, error) {
	var shortBuffer bytes.Buffer
	id := <-s.idChannel
	var err error
	for id > 0 {
		if id > ALPH_LAST {
			err = shortBuffer.WriteByte(ALPH[ALPH_LAST])
		} else {
			err = shortBuffer.WriteByte(ALPH[id])
		}
		id -= ALPH_LAST
		if err != nil {
			return nil, err
		}
	}

	return shortBuffer.Bytes(), nil
}

// Структура хэндлера
type Shorter struct {
	store Store
}

func (sh *Shorter) Init(db *bolt.DB) error {
	sh.store = Store{}
	err := sh.store.Init(db)
	return err
}

// Обработка POST-запросов
// Если ссылка уже есть в БД - возвращется оттуда, иначе генерируется
func (sh *Shorter) doPost(ctx *fasthttp.RequestCtx) {
	fullURL := ctx.PostBody()
	if len(fullURL) == 0 {
		ctx.Error("Body is empty", fasthttp.StatusBadRequest)
	}
	short, err := sh.store.Save(fullURL)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	fmt.Fprintf(ctx, "Short url: %s/%s", ctx.Host(), short)
	return
}

// Обработка GET-запросов
// Если ссылка уже есть в БД - возвращется текстом, иначе возвращается 404
func (sh *Shorter) doGet(ctx *fasthttp.RequestCtx) {
	short := ctx.Path()[1:]
	if full, err := sh.store.getFull(short); full != nil {
		// ctx.Redirect(full, fasthttp.StatusMovedPermanently)
		fmt.Fprintf(ctx, "Full url: %s", full)
	} else if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else {
		ctx.NotFound()
	}
	return
}

func (sh *Shorter) doDefault(ctx *fasthttp.RequestCtx) {
	ctx.Error("Method not allowed!", fasthttp.StatusMethodNotAllowed)
}

// Рутер HTTP-методов
func (sh *Shorter) Router(ctx *fasthttp.RequestCtx) {
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

func main() {
	conf := Configuration{}
	if _, err := toml.DecodeFile(os.Args[1], &conf); err != nil {
		panic(err)
	}

	db, err := bolt.Open(conf.DBPath, 0600, nil)
	if err != nil {
		panic(err)
	}

	shorter := Shorter{}
	shorter.Init(db)
	fmt.Printf("Server run on %s\n", conf.Address)
	fasthttp.ListenAndServe(conf.Address, shorter.Router)

	shorter.store.Close()
	db.Close()
}
