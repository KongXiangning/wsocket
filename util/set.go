package util

import (
	"sync"
	"time"
)

type Set struct {
	m map[interface{}]int64
	sync.RWMutex
}

func New() *Set {
	return &Set{
		m: map[interface{}]int64{},
	}
}

func (s *Set) Add(item interface{}) bool {
	if s.Has(item) {
		return false
	} else {
		s.Lock()
		defer s.Unlock()
		s.m[item] = time.Now().Unix()
		return true
	}
}

func (s *Set) Remove(item interface{}) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, item)
}

func (s *Set) Has(item interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[item]
	return ok
}
