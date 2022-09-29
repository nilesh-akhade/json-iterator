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

// Remove and return top element of stack. Return false if stack is empty.
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
	isDelim bool
	delim   json.Delim
	name    string
}

func (j *jsonToken) String() string {
	if j.isDelim {
		return j.delim.String()
	}
	return j.name
}

func (j *jsonToken) IsDelim() bool {
	return j.isDelim
}

func (j *jsonToken) IsArrayStart() bool {
	return j.IsDelim() && j.String() == "["
}

func (j *jsonToken) IsArrayEnd() bool {
	return j.IsDelim() && j.String() == "]"
}

func (j *jsonToken) IsObjStart() bool {
	return j.IsDelim() && j.String() == "{"
}

func (j *jsonToken) IsObjEnd() bool {
	return j.IsDelim() && j.String() == "}"
}

func (j *jsonToken) IsName() bool {
	return !j.isDelim
}

func NewDelimJsonToken(delim json.Delim) jsonToken {
	return jsonToken{
		isDelim: true,
		delim:   delim,
	}
}

func NewNameJsonToken(name string) jsonToken {
	return jsonToken{
		name: name,
	}
}
