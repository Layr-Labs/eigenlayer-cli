package common

import (
	"fmt"
	"os"
	"path"

	"github.com/gocarina/gocsv"
)

func WriteToJSON(data []byte, filePath string) error {
	dir := path.Dir(filePath)
	// Ensure the directory exists
	err := ensureDir(dir)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

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
	dir := path.Dir(filePath)
	// Ensure the directory exists
	err := ensureDir(dir)
	if err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

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

// Ensure that the directory exists, creating it if necessary
func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
