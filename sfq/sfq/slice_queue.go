package sfq

import (
	"errors"
	"sync"
)

var _ Queue[struct{}] = &sliceQueue[struct{}]{}

var ErrQueueFull = errors.New("queue is full")

type sliceQueue[T any] struct {
	items   []T
	maxSize int
	count   int
	headIdx int

	mu sync.Mutex
}

func NewSliceQueue[T any](maxSize int) Queue[T] {
	initialCap := maxSize
	if initialCap <= 0 || initialCap > 64 {
		initialCap = 64
	}

	return &sliceQueue[T]{
		items:   make([]T, initialCap),
		maxSize: maxSize,
	}
}

func (q *sliceQueue[T]) Push(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.maxSize > 0 && q.count >= q.maxSize {
		return ErrQueueFull
	}

	if q.count == len(q.items) {
		q.grow()
	}

	tail := (q.headIdx + q.count) % len(q.items)
	q.items[tail] = item
	q.count++

	return nil
}

func (q *sliceQueue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T
	if q.count == 0 {
		return zero, false
	}

	item := q.items[q.headIdx]
	q.items[q.headIdx] = zero
	q.headIdx = (q.headIdx + 1) % len(q.items)
	q.count--

	return item, true
}

func (q *sliceQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return int(q.count)
}

func (q *sliceQueue[T]) grow() {
	newCap := len(q.items) * 2
	if newCap == 0 {
		newCap = 8
	}

	newItems := make([]T, newCap)
	for i := 0; i < q.count; i++ {
		newItems[i] = q.items[(q.headIdx+i)%len(q.items)]
	}

	q.items = newItems
	q.headIdx = 0
}
