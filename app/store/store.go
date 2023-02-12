package store

import "0xgirish.eth/redis/app/resp/types"

type Store struct {
	cache map[types.BulkString]types.BulkString
}

func (s *Store) SET(key, value types.BulkString) {
	s.cache[key] = value
}

func (s *Store) GET(key types.BulkString) types.BulkString {
	return s.cache[key]
}

func (s *Store) DEL(key types.BulkString) types.Int {
	if _, ok := s.cache[key]; ok {
		delete(s.cache, key)
		return types.Int(1)

	}

	return types.Int(0)
}

func New() *Store {
	return &Store{
		cache: map[types.BulkString]types.BulkString{},
	}
}