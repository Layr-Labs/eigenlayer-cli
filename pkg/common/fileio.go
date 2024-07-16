package common

import (
	"fmt"
	"os"
	"path"

	"github.com/gocarina/gocsv"
)

func WriteToJSON(data []byte, filePath string) error {
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

func WriteToCSV(data interface{}, filePath string) error {
	file, err := os.Create(path.Clean(filePath))
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	err = gocsv.MarshalFile(data, file)
	if err != nil {
		return err
	}

	return nil
}
