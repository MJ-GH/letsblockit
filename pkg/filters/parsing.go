package filters

import (
	"bytes"
	"errors"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

var validate = validator.New()


func ParseFilter(name string, reader io.Reader) (*Filter, error) {
	filter := &Filter{
		Name: name,
	}
	return filter, parse(reader, filter)
}

func parseFilterAndTest(name string, reader io.Reader) (*filterAndTests, error) {
	filter := &filterAndTests{
		Filter: Filter{
			Name: name,
		},
	}
	return filter, parse(reader, filter)
}

func parse(reader io.Reader, filter filter) error {
	// Read the whole input file and parse the YAML block
	input, err := io.ReadAll(reader)
	err = yaml.Unmarshal(input, filter)
	if err != nil {
		return err
	}

	// Find the separator and parse the markdown after it
	pos := bytes.Index(input, yamlSeparator)
	if pos < 0 {
		return errors.New("separator not found")
	}
	pos += len(yamlSeparator)
	pos += bytes.Index(input[pos:], newLine)
	filter.SetDescription(blackfriday.Run(input[pos:]))

	// Run input validation
	return validate.Struct(filter)
}
