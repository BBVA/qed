package common

type TreeNavigator interface {
	Root() Position
	IsLeaf(Position) bool
	IsRoot(Position) bool
	GoToLeft(Position) Position
	GoToRight(Position) Position
	DescendToFirst(Position) Position
	DescendToLast(Position) Position
}
