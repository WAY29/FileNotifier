package parse

import (
	"os"

	"github.com/WAY29/FileNotifier/internal/structs"
	"gopkg.in/yaml.v2"
)

func ParseTemplate(filename string) (*structs.Template, error) {
	template := &structs.Template{}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err != nil {
		return nil, err
	}

	err = yaml.NewDecoder(f).Decode(template)

	if err != nil {
		return nil, err
	}

	return template, nil
}
