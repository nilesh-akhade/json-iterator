package jsoniterator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func NewJsonRecordsIterator(filename string) Iterator {
	return &iterator{
		filename:     filename,
		typeDecoders: make(map[string]TypeDecoder),
	}
}

type iterator struct {
	filename string
	file     *os.File

	stack   Stack
	decoder *json.Decoder

	stateVars    map[string]interface{}
	typeDecoders map[string]TypeDecoder

	err      error
	nextElem interface{}
}

func (i *iterator) Err() error {
	if i.err == io.EOF {
		return nil
	}
	return i.err
}

func (i *iterator) HasNext() bool {
	if i.file == nil {
		err := i.start()
		if err != nil {
			i.err = err
			return false
		}
		i.nextElem, err = i.next()
		i.err = err
	}
	return i.err == nil && i.nextElem != nil
}

func (i *iterator) Next() interface{} {
	var err error
	if i.file == nil {
		err = i.start()
		if err != nil {
			i.file.Close()
			i.err = err
			return nil
		}

		firstElem, err := i.next()
		if err != nil {
			i.file.Close()
			i.err = err
			return nil
		}
		if firstElem != nil {
			i.nextElem, err = i.next()
			if err != nil {
				i.file.Close()
				i.err = err
			}
		}
		return firstElem
	}
	current := i.nextElem
	if current != nil {
		i.nextElem, err = i.next()
		if err != nil {
			i.file.Close()
			i.err = err
		}
	} else {
		i.file.Close()
	}
	return current
}

func (i *iterator) start() error {
	var err error
	i.file, err = os.Open(i.filename)
	if err != nil {
		return fmt.Errorf("error opening file: %s, %w", i.filename, err)
	}
	reader := bufio.NewReader(i.file)
	i.decoder = json.NewDecoder(reader)
	i.stateVars = map[string]interface{}{}
	return nil
}

func (i *iterator) RegisterTypeDecoder(jsonPath string, decoder TypeDecoder) {
	i.typeDecoders[jsonPath] = decoder
}

func (i *iterator) next() (interface{}, error) {
	var ok bool
	var tokenFromStack jsonToken
	jsonPath := i.stack.String()
	if d, found := i.typeDecoders[jsonPath]; found {
		if i.decoder.More() {
			return d(i.decoder, i.stateVars)
		}
		i.decoder.Token() // ]
		i.stack.Pop()
	}

	for {
		jsonPath = i.stack.String()
		t, err := i.decoder.Token()
		if err == io.EOF {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		switch t.(type) {
		case json.Delim:
			token := NewJsonToken(t)
			switch {
			case token.IsObjStart():
				i.stack.Push(token)
			case token.IsObjEnd():
				tokenFromStack, ok = i.stack.Pop()
				if !ok {
					return nil, fmt.Errorf("expected { before }")
				}
				if tokenFromStack.IsObjStart() {
					if i.stack.IsEmpty() {
						continue // EOF
					}
					tokenFromStack = i.stack.Peek()
					if tokenFromStack.IsName() {
						i.stack.Pop()
					}
					i.deleteStateVars(jsonPath)
				}
			case token.IsArrayStart():
				if d, found := i.typeDecoders[jsonPath]; found {
					if i.decoder.More() {
						return d(i.decoder, i.stateVars)
					}
					i.decoder.Token() // ]
					i.stack.Pop()     // array name
				} else {
					i.stack.Push(token)
				}
			case token.IsArrayEnd():
				i.stack.Pop() // [
				i.stack.Pop() // array name
				// Remove stateVars because we are moving out of scope
				i.deleteStateVars(jsonPath)
			}
			continue
		case string:
			// t is name or val?
			tokenFromStack = i.stack.Peek()
			if tokenFromStack.IsName() {
				// t is a json value
				i.stack.Pop()
				i.stateVars[jsonPath] = t
			} else {
				// t is a json name
				token := NewJsonToken(t)
				if d, found := i.typeDecoders["{}"+jsonPath+"."+token.String()]; found {
					if i.decoder.More() {
						return d(i.decoder, i.stateVars)
					}
				} else {
					i.stack.Push(token)
				}
			}
		default: // bool, floats, null and other json literals
			i.stack.Pop()
			i.stateVars[jsonPath] = t
		}

	}
}

func (i *iterator) deleteStateVars(jsonPath string) {
	for key := range i.stateVars {
		if strings.HasPrefix(key, jsonPath) {
			delete(i.stateVars, key)
		}
	}
}
