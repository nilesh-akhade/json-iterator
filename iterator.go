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
				return nil
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
			return nil
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
	var jt jsonToken
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
		if jd, ok := t.(json.Delim); ok {
			strJD := jd.String()
			switch strJD {
			case "{":
				i.stack.Push(NewJsonToken(jd))
			case "}":
				jt, _ = i.stack.Pop()
				if jt.IsObjStart() {
					if i.stack.IsEmpty() {
						continue
					}
					jt = i.stack.Peek()
					if jt.IsName() {
						i.stack.Pop()
					}
				}
			case "[":
				// fmt.Println("+ARR_START")
				if d, found := i.typeDecoders[jsonPath]; found {
					if i.decoder.More() {
						return d(i.decoder, i.stateVars)
					}
					i.decoder.Token() // ]
					i.stack.Pop()     // Misconfig
				} else {
					i.stack.Push(NewJsonToken(jd))
				}
			case "]":
				i.stack.Pop()                 // [
				nameOfArr, _ := i.stack.Pop() // name of array
				nameOfArr.String()
				// fmt.Printf("-ARR_END=%v\n", nameOfArr.String())
				// Remove stateVars because we are moving out of scope
				for key := range i.stateVars {
					if strings.HasPrefix(key, jsonPath) {
						// fmt.Println("removed", key)
						delete(i.stateVars, key)
					}
				}
			}
			continue
		}

		tokenFromStack := i.stack.Peek()
		if tokenFromStack.IsName() {
			// t is a json value
			i.stack.Pop()
			i.stateVars[jsonPath] = t
		} else {
			tokenRead := NewJsonToken(t)
			if tokenRead.IsName() {
				i.stack.Push(tokenRead)
			}
		}
	}
}
