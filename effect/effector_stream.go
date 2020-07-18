package effect

// EffectorStream is the interface for effector stream.
type EffectorStream interface {
	Next() bool
	Explain() []int
	PushEffect(eft Effect) bool
}

