package effect

// EffectorStream is the interface for effector stream.
type EffectorStream interface {
	Current() bool
	Explain() []uint
	PushEffect(eft Effect) bool
}