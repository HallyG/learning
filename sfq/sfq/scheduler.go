package sfq

type Keyer interface {
	Key() string
}

type Scheduler[T Keyer] interface {
	Enqueue(item T)
	Dequeue() (T, bool)
	Perturb()
}
