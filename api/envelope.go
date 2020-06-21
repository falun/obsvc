package api

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Envelope interface {
	WithHttpStatus(int) Envelope

	HttpStatus() int
	EncodeJson(pretty bool) []byte
	EncodeYaml() []byte
}

type envelopeImpl struct {
	httpStatus *int
	Error      bool        `json:"error" yaml:"error"`
	ErrorCode  *string     `json:"error_code,omitempty" yaml:"error_code,omitempty"`
	Result     interface{} `json:"result,omitempty" yaml:"result,omitempty"`
}

func NewResponse(data interface{}) Envelope {
	return &envelopeImpl{Error: false, Result: data}
}

func NewError(code string) Envelope {
	var s *string
	if code != "" {
		s = &code
	}

	return &envelopeImpl{
		Error:     true,
		ErrorCode: s,
		Result:    nil,
	}
}

func (e *envelopeImpl) HttpStatus() int {
	if e.httpStatus != nil {
		return *e.httpStatus
	}
	if e.Error {
		return 500
	}
	return 200
}

func (e *envelopeImpl) EncodeJson(pretty bool) []byte {
	var b []byte
	var err error

	if pretty {
		b, err = json.MarshalIndent(e, "", "  ")
	} else {
		b, err = json.Marshal(e)
	}

	if err != nil {
		b = []byte(fmt.Sprintf(`{ "error": true, "result": %q }`, fmt.Sprintf("%v", err)))
	}

	return b
}

func (e *envelopeImpl) EncodeYaml() []byte {
	b, err := yaml.Marshal(e)
	if err != nil {
		b = []byte(fmt.Sprintf(
			`error: true
result: %q`, fmt.Sprintf("%v", err)))
	}

	return b
}

func (e *envelopeImpl) WithHttpStatus(code int) Envelope {
	e.httpStatus = &code
	return e
}

func (e *envelopeImpl) WithExtras(extras interface{}) Envelope {
	e.Result = extras
	return e
}
