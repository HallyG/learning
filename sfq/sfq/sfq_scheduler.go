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
	queues []Queue[T]

	cursor       int
	perturbation uint64
	subsets      int

	mu sync.RWMutex
}

func NewSFQScheduler[T Keyer](queue ...Queue[T]) Scheduler[T] {
	s := &sfqScheduler[T]{
		queues:  queue,
		subsets: 1,
	}

	if s.subsets > len(queue) {
		s.subsets = len(queue)
	}

	return s
}

func (s *sfqScheduler[T]) Enqueue(item T) error {
	s.mu.RLock()
	perturbation := s.perturbation
	s.mu.RUnlock()

	queueIdx := s.bestQueue(item.Key(), perturbation)
	if err := s.queues[queueIdx].Push(item); err != nil {
		return fmt.Errorf("enqueueing item: %w", err)
	}

	return nil
}

func (s *sfqScheduler[T]) bestQueue(key string, perturbation uint64) int {
	result := make([]int, 0, s.subsets)
	seen := make(map[int]struct{})

	var pBuf [8]byte
	binary.BigEndian.PutUint64(pBuf[:], perturbation)

	for i := uint64(0); len(result) < int(s.subsets); i++ {
		buf := make([]byte, 0, len(key)+16)
		buf = append(buf, key...)
		buf = append(buf, pBuf[:]...)
		buf = binary.BigEndian.AppendUint64(buf, i)

		sum := sha256.Sum256(buf)
		hash := binary.BigEndian.Uint64(sum[:])

		queueIdx := int(hash % uint64(len(s.queues))) //nolint:gosec
		if _, ok := seen[queueIdx]; ok {
			continue
		}

		seen[queueIdx] = struct{}{}
		result = append(result, queueIdx)
	}

	best := result[0]
	for _, idx := range result[1:] {
		if s.queues[idx].Len() < s.queues[best].Len() {
			best = idx
		}
	}

	return best
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

	s.perturbation++

}
