package jsoniterator

import "encoding/json"

// TypeDecoder callback allows to read custom type from `json.Decoder`
//
// Values read earlier by iterator are available as a stateVariable map
type TypeDecoder func(*json.Decoder, map[string]interface{}) (interface{}, error)

// Iterator provides an interface to iterate over nested json objects
// The path of nested types has to be registered using `RegisterTypeDecoder`
//
//	iter := NewIterator(filename)
type Iterator interface {
	HasNext() bool
	Next() interface{}
	Err() error
	RegisterTypeDecoder(jsonPath string, decoder TypeDecoder)
}
