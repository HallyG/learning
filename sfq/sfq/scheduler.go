package sfq

type Keyer interface {
	Key() string
}

type Scheduler[T Keyer] interface {
	Enqueue(item T) error
	Dequeue() (T, bool)
	Perturb()
}
