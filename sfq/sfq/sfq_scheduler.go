package sfq

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
)

var _ Scheduler[*Request] = &sfqScheduler[*Request]{}

type Request struct {
	ID string
}

func (r *Request) Key() string {
	return r.ID
}

type sfqScheduler[T Keyer] struct {
	queues      []Queue[T]
	cursor      int
	pertubation int

	mu sync.RWMutex
}

func NewSFQScheduler[T Keyer](queue ...Queue[T]) Scheduler[T] {
	return &sfqScheduler[T]{
		queues: queue,
	}
}

func (s *sfqScheduler[T]) Enqueue(item T) error {
	s.mu.RLock()
	pertubation := s.pertubation
	s.mu.RUnlock()

	hash := sha256.Sum256(append([]byte(item.Key()), byte(pertubation)))
	hashNum := int(binary.BigEndian.Uint32(hash[:]))

	queueIdx := hashNum % len(s.queues)
	s.queues[queueIdx].Push(item)

	fmt.Println(item.Key(), queueIdx)

	return nil
}

func (s *sfqScheduler[T]) Dequeue() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.queues)
	var zero T
	if n == 0 {
		return zero, false
	}

	for i := range n {
		idx := (s.cursor + i) % n

		if item, ok := s.queues[idx].Pop(); ok {
			s.cursor = (idx + 1) % n
			return item, true
		}
	}

	return zero, false
}

func (s *sfqScheduler[T]) Perturb() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pertubation++

}
