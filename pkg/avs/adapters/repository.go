package adapters

import (
	"plugin"
)

type Repository interface {
	LoadPlugin(name string, url string) (*plugin.Plugin, error)
	LoadResource(name string, resource string) ([]byte, error)
}
