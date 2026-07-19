package sfq

type Queue[T any] interface {
	Push(item T) error
	Pop() (T, bool)
	Len() int
}
