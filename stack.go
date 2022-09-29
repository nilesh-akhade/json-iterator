package jsoniterator

import (
	"encoding/json"
	"strings"
)

type Stack []jsonToken

// IsEmpty: check if stack is empty
func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

// String: returns items separated by dot
func (s *Stack) String() string {
	if len(*s) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString((*s)[0].String())
	for _, s := range (*s)[1:] {
		b.WriteString(".")
		b.WriteString(s.String())
	}
	return b.String()
}

// Push a new value onto the stack
func (s *Stack) Push(str jsonToken) {
	*s = append(*s, str)
}

// Peek a value without pop
func (s *Stack) Peek() jsonToken {
	index := len(*s) - 1
	return (*s)[index]
}

// Pop Removes and returns top element of the stack. Return false if stack is empty.
func (s *Stack) Pop() (jsonToken, bool) {
	if s.IsEmpty() {
		return jsonToken{}, false
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		*s = (*s)[:index]
		return element, true
	}
}

type jsonToken struct {
	json.Token
}

const (
	objStart   = json.Delim('{')
	objEnd     = json.Delim('}')
	arrayStart = json.Delim('[')
	arrayEnd   = json.Delim(']')
)

func (j *jsonToken) String() string {
	switch j.Token.(type) {
	case string:
		return j.Token.(string)
	case json.Delim:
		return j.Token.(json.Delim).String()
	}
	return ""
}

func (j *jsonToken) IsDelim() bool {
	_, ok := j.Token.(json.Delim)
	return ok
}

func (j *jsonToken) checkDelim(delim json.Delim) bool {
	if v, ok := j.Token.(json.Delim); ok {
		return v == delim
	}
	return false
}

func (j *jsonToken) IsArrayStart() bool {
	return j.checkDelim(arrayStart)
}

func (j *jsonToken) IsArrayEnd() bool {
	return j.checkDelim(arrayEnd)
}

func (j *jsonToken) IsObjStart() bool {
	return j.checkDelim(objStart)
}

func (j *jsonToken) IsObjEnd() bool {
	return j.checkDelim(objEnd)
}

func (j *jsonToken) IsName() bool {
	_, ok := j.Token.(string)
	return ok
}

func NewJsonToken(token json.Token) jsonToken {
	// TODO: Do not allow numbers, null
	return jsonToken{token}
}
