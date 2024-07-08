package keys

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKeyPath(t *testing.T) {
	t.Skip("Skip test")
	homePath, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		keyType      string
		keyPath      string
		keyName      string
		err          error
		expectedPath string
	}{
		{
			name:         "correct key path using keyname",
			keyType:      KeyTypeECDSA,
			keyName:      "test",
			err:          nil,
			expectedPath: filepath.Join(homePath, OperatorKeystoreSubFolder, "test.ecdsa.key.json"),
		},
		{
			name:         "correct key path using keypath",
			keyType:      KeyTypeECDSA,
			keyPath:      filepath.Join(homePath, "x.json"),
			err:          nil,
			expectedPath: filepath.Join(homePath, "x.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := getKeyPath(tt.keyPath, tt.keyName, tt.keyType)
			if err != nil {
				t.Fatal(err)
			}

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}
