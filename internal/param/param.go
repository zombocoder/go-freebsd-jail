package param

// Param represents a single jail parameter to pass to the C layer.
type Param struct {
	Name   string
	Value  string
	IsBool bool
}
