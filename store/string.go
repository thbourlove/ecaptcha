package store

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

type StringStore struct {
	conn redis.Conn
}

func (r *StringStore) Get(key string, clear bool) (value string) {
	value, err := redis.String(r.conn.Do("GET", key))
	if err != nil {
		log.Println(err)
		return ""
	}
	if clear {
		_, err = r.conn.Do("DEL", key)
		if err != nil {
			log.Println(err)
		}
	}
	return value
}

func (r *StringStore) Set(key string, value string) {
	_, err := r.conn.Do("SET", key, value)
	if err != nil {
		log.Println(err)
	}
}

func NewStringStore(c redis.Conn) *StringStore {
	return &StringStore{c}
}
