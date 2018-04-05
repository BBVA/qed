package history

type Store interface {
	Add(node *Node) error
	Get(*Position) (*Node, error)
}
