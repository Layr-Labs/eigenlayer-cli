package plugin

type PluginConfiguration interface {
	Set(key string, value interface{})
}

type PluginSpecification interface {
	Validate() error
}

type PluginCoordinator interface {
	Register() error
	OptIn() error
	OptOut() error
	Deregister() error
	Status() (int, error)
}
