package jsoniterator

import "encoding/json"

type GoTypeJsonDecoder func(*json.Decoder, map[string]interface{}) (interface{}, error)

type JsonRecordsIterator interface {
	HasNext() bool
	Next() interface{}
	Err() error
	RegisterGoTypeDecoder(jsonPath string, d GoTypeJsonDecoder)
}
