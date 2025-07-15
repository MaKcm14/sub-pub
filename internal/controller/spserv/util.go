package spserv

import (
	"sync"

	"github.com/MaKcm14/vk-test/internal/controller/spserv/sprpc"
)

// SPServer defines the common interface for every Sub/Pub server implementation.
type SPServer interface {
	sprpc.PubSubServer
	Close()
}

type syncMap struct {
	m   map[int]chan string
	rwm sync.RWMutex
}

func newSyncMap() syncMap {
	return syncMap{
		m: make(map[int]chan string),
	}
}

func (s *syncMap) Add(key int, ch chan string) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	s.m[key] = ch
}

func (s *syncMap) Delete(key int) chan string {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	ch := s.m[key]
	delete(s.m, key)

	return ch
}

func (s *syncMap) Range(f func(ch chan string)) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	for _, ch := range s.m {
		f(ch)
	}
}
