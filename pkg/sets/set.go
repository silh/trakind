package sets

import (
	"golang.org/x/exp/constraints"
	"sync"
)

var empty = struct{}{}

type Action[F constraints.Ordered] func(item F)

type Set[F constraints.Ordered] interface {
	Add(item F)
	Remove(item F)
	Contains(item F) bool
	ForEach(action Action[F])
}

type HashSet[F constraints.Ordered] struct {
	container map[F]struct{}
}

func New[F constraints.Ordered]() *HashSet[F] {
	return &HashSet[F]{container: make(map[F]struct{})}
}

func (s *HashSet[F]) Add(item F) {
	s.container[item] = empty
}

func (s *HashSet[F]) Remove(item F) {
	delete(s.container, item)
}

func (s *HashSet[F]) Contains(item F) bool {
	_, ok := s.container[item]
	return ok
}

func (s *HashSet[F]) ForEach(action Action[F]) {
	for k := range s.container {
		action(k)
	}
}

type ConcurrentHashSet[F constraints.Ordered] struct {
	set  Set[F]
	lock sync.RWMutex
}

func NewConcurrent[F constraints.Ordered]() *ConcurrentHashSet[F] {
	return &ConcurrentHashSet[F]{
		set:  New[F](),
		lock: sync.RWMutex{},
	}
}

func (s *ConcurrentHashSet[F]) Add(item F) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set.Add(item)
}

func (s *ConcurrentHashSet[F]) Remove(item F) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.set.Remove(item)
}

func (s *ConcurrentHashSet[F]) Contains(item F) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.set.Contains(item)
}

func (s *ConcurrentHashSet[F]) ForEach(action Action[F]) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	s.set.ForEach(action)
}
