package store

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

type BytesStore struct {
	conn redis.Conn
}

func (r *BytesStore) Get(key string, clear bool) (value []byte) {
	value, err := redis.Bytes(r.conn.Do("GET", key))
	if err != nil {
		log.Println(err)
	}
	if clear {
		_, err = r.conn.Do("DEL", key)
		if err != nil {
			log.Println(err)
		}
	}
	return value
}

func (r *BytesStore) Set(key string, value []byte) {
	_, err := r.conn.Do("SET", key, value)
	if err != nil {
		log.Println(err)
	}
}

func NewBytesStore(c redis.Conn) *BytesStore {
	return &BytesStore{c}
}
