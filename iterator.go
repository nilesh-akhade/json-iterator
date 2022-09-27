package jsoniterator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func NewJsonRecordsIterator(filename string) JsonRecordsIterator {
	return &jsonRecordsIterator{
		filename:       filename,
		goTypeDecoders: make(map[string]GoTypeJsonDecoder),
	}
}

type jsonRecordsIterator struct {
	filename string
	file     *os.File

	stack   Stack
	decoder *json.Decoder

	stateVars      map[string]interface{}
	goTypeDecoders map[string]GoTypeJsonDecoder

	err      error
	nextElem interface{}
}

func (i *jsonRecordsIterator) Err() error {
	return i.err
}

func (i *jsonRecordsIterator) HasNext() bool {
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

func (i *jsonRecordsIterator) Next() interface{} {
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

func (i *jsonRecordsIterator) start() error {
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

func (i *jsonRecordsIterator) RegisterGoTypeDecoder(jsonPath string, d GoTypeJsonDecoder) {
	i.goTypeDecoders[jsonPath] = d
}

func (i *jsonRecordsIterator) next() (interface{}, error) {
	var jt jsonToken
	jsonPath := i.stack.String()
	if d, found := i.goTypeDecoders[jsonPath]; found {
		if i.decoder.More() {
			return d(i.decoder, i.stateVars)
		}
		// fmt.Printf("-ARR_UNMARSHAL_END\n")
		i.decoder.Token() // ]
		i.stack.Pop()     // Misconfig
	}

	for {
		jsonPath = i.stack.String()
		t, err := i.decoder.Token()
		if err == io.EOF {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if jd, ok := t.(json.Delim); ok {
			strJD := jd.String()
			switch strJD {
			case "{":
				// fmt.Printf("+OBJ_START\n")
				i.stack.Push(NewDelimJsonToken(jd))
			case "}":
				jt, _ = i.stack.Pop()
				if jt.IsDelim() && jt.String() == "{" {
					if i.stack.IsEmpty() {
						continue
					}
					jt = i.stack.Peek()
					if jt.IsDelim() && jt.String() == "[" {
						// fmt.Printf("-ARRAY_ELEM_OBJ_END\n")
					} else if jt.IsName() {
						// fmt.Printf("-OBJ_END=%v\n", jt.String())
						i.stack.Pop()
					}
				}
			case "[":
				// fmt.Println("+ARR_START")
				if d, found := i.goTypeDecoders[jsonPath]; found {
					if i.decoder.More() {
						return d(i.decoder, i.stateVars)
					}
					// fmt.Printf("-ARR_UNMARSHAL_END\n")
					i.decoder.Token() // ]
					i.stack.Pop()     // Misconfig
				} else {
					i.stack.Push(NewDelimJsonToken(jd))
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

		jt = i.stack.Peek()
		if jt.IsName() {
			// t is a json value
			jKey, _ := i.stack.Pop()
			i.stateVars[jsonPath] = t
			jKey.String()
			// fmt.Printf("%v=%v\n", jKey.String(), t)
		} else if jt.IsDelim() && jt.String() == "[" {
			// t is a array elem
			// fmt.Println("-ARR_ELEM=", t)
		} else {
			// t is a json name
			i.stack.Push(NewNameJsonToken(t.(string)))
			// fmt.Println("key:", t.(string))
		}
	}
}
