package jsoniterator

import "encoding/json"

// TypeDecoder use .Decode() to read custom type
type TypeDecoder func(*json.Decoder, map[string]interface{}) (interface{}, error)

type Iterator interface {
	HasNext() bool
	Next() interface{}
	Err() error
	RegisterTypeDecoder(jsonPath string, decoder TypeDecoder)
}
