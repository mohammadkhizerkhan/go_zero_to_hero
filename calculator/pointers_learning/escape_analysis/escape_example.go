package pointerslearning

// Sink variables prevent compiler optimizations from removing benchmark work.
var SinkInt int
var SinkPtr *int

// StackOnly returns a plain int, so x can stay on the stack.
func StackOnly() int {
	x := 42
	return x
}

// HeapEscape returns the address of a local variable, so x must escape.
func HeapEscape() *int {
	x := 42
	return &x
}
