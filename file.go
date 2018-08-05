package main

import (
	"os"
)

func NewFile(filename string) (*os.File, error) {

	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}