package common

type Signer interface {
	Sign(data []byte) ([]byte, error)
}
