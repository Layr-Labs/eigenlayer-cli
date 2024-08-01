package common

type OutputType string

const (
	// MaxAddressLength Magic number 42 is the max length of an address.
	// But it's also answer to the life, universe and everything.
	MaxAddressLength = 42

	OutputType_Calldata OutputType = "calldata"
	OutputType_Pretty   OutputType = "pretty"
	OutputType_Json     OutputType = "json"
)
