package main

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type yamlReader struct {
	file *os.File
}

func newYamlReader() *yamlReader {
	return &yamlReader{}
}

func (yr *yamlReader) fromFile(filePath string) *yamlReader {
	file, err := os.Open(filePath)
	if err != nil {
		if pathError, ok := err.(*os.PathError); ok {
			if strings.HasSuffix(pathError.Path, ".yaml") {
				return yr.fromFile(strings.TrimSuffix(pathError.Path, ".yaml") + ".yml")
			}
			log.Fatalf("error while opening provided yaml file: %v", err)
		}
	}

	yr.file = file
	return yr
}

func (yr *yamlReader) deserializeInto(data any) error {
	return yaml.NewDecoder(yr.file).Decode(data)
}
