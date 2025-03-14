package signers

type Web3Signer struct {
	endpoint      string
	signerAddress string
}

func (m *Web3Signer) Sign(data []byte) ([]byte, error) {
	panic("Web3 signer not implemented")
}

func NewWeb3Signer(endpoint string, signerAddress string) (*Web3Signer, error) {
	return &Web3Signer{endpoint: endpoint, signerAddress: signerAddress}, nil
}
