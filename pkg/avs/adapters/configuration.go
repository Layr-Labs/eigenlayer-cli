package adapters

type Configuration interface {
	Get(key string) (interface{}, error)
	GetAll() map[string]interface{}
	Prompt(key string, required bool, hidden bool) (interface{}, error)
	Unmarshal(config any) error
}
