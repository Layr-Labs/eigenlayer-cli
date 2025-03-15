package types

type Web3SignerConfig struct {
	Url string `yaml:"url" json:"url"`
}

func (w *Web3SignerConfig) MarshalYAML() (interface{}, error) {
	return struct {
		Url string `yaml:"url"`
	}{
		Url: w.Url,
	}, nil
}

func (w *Web3SignerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		Url string `yaml:"url"`
	}
	if err := unmarshal(&aux); err != nil {
		return err
	}
	w.Url = aux.Url
	return nil
}
