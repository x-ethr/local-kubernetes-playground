package logging

// Flag represents a given logging Option's value(s).
type Flag[Value interface{}] interface {
	// Set will configure the given Setting implementation's logging flag.
	Set(v Value)

	// Value returns the implementation's value reference.
	Value() Value
}

type flag[Value interface{}] struct {
	Flag[Value]

	v Value // v represents the option's user-defined value.
}

func (f *flag[Value]) Set(v Value) {
	f.v = v
}

func (f *flag[Value]) Value() Value {
	return f.v
}
