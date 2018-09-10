package store

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/boltdb/bolt"
)

const (
	alphabet          = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphabetLastIndex = int64(len(alphabet) - 1)
)

type Store struct {
	lastID    int64      // Последний ID для генерации хэша
	idChannel chan int64 // который отдает ID для гарантированой инкрментации в одной точке
	db        *bolt.DB
	wg        sync.WaitGroup
}

// Init - initialize storage, read properies from DB, start counnter gourutine to produce uniq ID
func (s *Store) Init(db *bolt.DB) error {
	s.idChannel = make(chan int64)
	s.db = db
	s.wg = sync.WaitGroup{}
	err := s.db.Batch(func(tx *bolt.Tx) error {
		var err error
		// shortToFull - mapping short URLs to full
		_, err = tx.CreateBucketIfNotExists([]byte("shortToFull"))
		if err != nil {
			panic(err)
		}
		// fullToShort - mapping full URLs to short
		_, err = tx.CreateBucketIfNotExists([]byte("fullToShort"))
		if err != nil {
			panic(err)
		}

		// Props - Properties bucket
		props := tx.Bucket([]byte("Props"))
		if props != nil {
			LastID := props.Get([]byte("LastID"))
			s.lastID = int64(binary.LittleEndian.Uint64(LastID)) - 1
		}

		return err
	})
	if err != nil {
		return err
	}
	// Produce incremental ID (by request from channel)
	go func(idChannel chan int64) {
		for {

			s.lastID++
			s.idChannel <- s.lastID

			go s.db.Batch(func(tx *bolt.Tx) error {

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

// Close - finish working with storage properly
func (s *Store) Close() {
	s.db.Close()
}

// Save - save full URL in database and returns hash for a short one
func (s *Store) Save(fullURL []byte) ([]byte, error) {

	if short, err := s.ShortURL(fullURL); short != nil {
		return short, nil
	} else if err != nil {
		return nil, err
	}

	short, err := s.Hash()
	if err != nil {
		return nil, err
	}

	// Saving data to database in background
	go s.db.Batch(func(tx *bolt.Tx) error {
		defer s.wg.Done()
		s.wg.Add(1)
		shortToFull := tx.Bucket([]byte("shortToFull"))
		fullToShort := tx.Bucket([]byte("fullToShort"))
		if err := shortToFull.Put(short, fullURL); err != nil {
			return err
		}
		if err := fullToShort.Put(fullURL, short); err != nil {
			return err
		}

		return nil
	})

	return short, err
}

// SaveLocked - same as Save() but waits untill date will be saved in database
func (s *Store) SaveLocked(fullURL []byte) ([]byte, error) {
	if short, err := s.ShortURL(fullURL); short != nil {
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
			return err
		}
		if err := fullToShort.Put(fullURL, short); err != nil {
			return err
		}
		return nil
	})

	return short, err
}

// FullURL - returns full URL by short URL
func (s *Store) FullURL(short []byte) ([]byte, error) {
	full := []byte{}
	err := s.db.View(func(tx *bolt.Tx) error {
		shortToFull := tx.Bucket([]byte("shortToFull"))
		full = shortToFull.Get(short)
		return nil
	})

	return full, err
}

// ShortURL - returns short URL by full URL
func (s *Store) ShortURL(full []byte) ([]byte, error) {
	short := []byte{}
	err := s.db.View(func(tx *bolt.Tx) error {
		fullToShort := tx.Bucket([]byte("fullToShort"))
		short = fullToShort.Get(full)
		return nil
	})

	return short, err
}

// Hash - produce incremental hash
func (s *Store) Hash() ([]byte, error) {
	var shortBuffer bytes.Buffer
	id := <-s.idChannel
	var err error
	index := int64(1)
	for id > 0 {
		index = id % alphabetLastIndex
		err = shortBuffer.WriteByte(alphabet[index])
		if err != nil {
			return nil, err
		}
		id /= alphabetLastIndex
	}

	return shortBuffer.Bytes(), nil
}
