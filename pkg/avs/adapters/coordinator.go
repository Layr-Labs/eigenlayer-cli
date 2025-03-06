package adapters

type Coordinator interface {
	Type() string
	Register() error
	OptIn() error
	OptOut() error
	Deregister() error
	Status() (int, error)
}
