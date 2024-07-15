package common

import (
	"fmt"
	"os"
	"path"
)

func WriteToJSON(data []byte, filePath string) error {
	// Write JSON data to a file
	file, err := os.Create(path.Clean(filePath))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}
