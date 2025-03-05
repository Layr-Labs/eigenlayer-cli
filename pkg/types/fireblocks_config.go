package types

type FireblocksConfig struct {
	APIKey           string `yaml:"api_key"            json:"api_key"`
	SecretKey        string `yaml:"secret_key"         json:"secret_key"`
	BaseUrl          string `yaml:"base_url"           json:"base_url"`
	VaultAccountName string `yaml:"vault_account_name" json:"vault_account_name"`

	SecretStorageType SecretStorageType `yaml:"secret_storage_type" json:"secret_storage_type"`

	// AWSRegion is the region where the secret is stored in AWS Secret Manager
	AWSRegion string `yaml:"aws_region" json:"aws_region"`

	// Timeout for API in seconds
	Timeout int64 `yaml:"timeout" json:"timeout"`
}

func (f *FireblocksConfig) MarshalYAML() (interface{}, error) {
	return struct {
		APIKey            string            `yaml:"api_key"`
		SecretKey         string            `yaml:"secret_key"`
		BaseUrl           string            `yaml:"base_url"`
		VaultAccountName  string            `yaml:"vault_account_name"`
		SecretStorageType SecretStorageType `yaml:"secret_storage_type"`
		AWSRegion         string            `yaml:"aws_region"`
		Timeout           int64             `yaml:"timeout"`
	}{
		APIKey:            f.APIKey,
		SecretKey:         f.SecretKey,
		BaseUrl:           f.BaseUrl,
		VaultAccountName:  f.VaultAccountName,
		SecretStorageType: f.SecretStorageType,
		AWSRegion:         f.AWSRegion,
		Timeout:           f.Timeout,
	}, nil
}

func (f *FireblocksConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var aux struct {
		APIKey            string            `yaml:"api_key"`
		SecretKey         string            `yaml:"secret_key"`
		BaseUrl           string            `yaml:"base_url"`
		VaultAccountName  string            `yaml:"vault_account_name"`
		SecretStorageType SecretStorageType `yaml:"secret_storage_type"`
		AWSRegion         string            `yaml:"aws_region"`
		Timeout           int64             `yaml:"timeout"`
	}
	if err := unmarshal(&aux); err != nil {
		return err
	}
	f.APIKey = aux.APIKey
	f.SecretKey = aux.SecretKey
	f.BaseUrl = aux.BaseUrl
	f.VaultAccountName = aux.VaultAccountName
	f.SecretStorageType = aux.SecretStorageType
	f.AWSRegion = aux.AWSRegion
	f.Timeout = aux.Timeout
	return nil
}
