package encoding

import (
	"fmt"
	"io"
)

type Encodable interface {
	Encode(writer io.Writer) error
	Decode(reader io.Reader) error
}

type encodeStep struct {
	name   string
	target Encodable
}

func step(name string, target Encodable) *encodeStep {
	return &encodeStep{name: name, target: target}
}

func encode(writer io.Writer, steps ...*encodeStep) error {
	for i := 0; i < len(steps); i++ {
		step := steps[i]
		err := step.target.Encode(writer)
		if err != nil {
			return fmt.Errorf("error encoding %s: %w", step.name, err)
		}
	}
	return nil
}

func decode(reader io.Reader, steps ...*encodeStep) error {
	for i := 0; i < len(steps); i++ {
		step := steps[i]
		err := step.target.Decode(reader)
		if err != nil {
			return fmt.Errorf("error encoding %s: %w", step.name, err)
		}
	}
	return nil
}
