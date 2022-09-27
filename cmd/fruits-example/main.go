package main

import (
	"encoding/json"
	"fmt"

	jsoniterator "github.com/nilesh-akhade/json-iterator"
)

func main() {
	iterator := jsoniterator.NewJsonRecordsIterator("fruits.json")
	iterator.RegisterGoTypeDecoder("{.trees.[.{.fruits", func(d *json.Decoder, m map[string]interface{}) (interface{}, error) {
		var fruit Fruit
		err := d.Decode(&fruit)
		if val, found := m["{.trees.[.{.id"]; found {
			if strVal, ok := val.(string); ok {
				fruit.TreeID = strVal
			}
		}
		return fruit, err
	})
	for iterator.HasNext() {
		item := iterator.Next()
		if fruit, ok := item.(Fruit); ok {
			fmt.Printf("%#v\n", fruit)
		}
	}
	if iterator.Err() != nil {
		fmt.Println(iterator.Err())
	}
}

type Fruit struct {
	TreeID string
	Name   string `json:"name,omitempty"`
	Color  string `json:"color,omitempty"`
	Taste  string `json:"taste,omitempty"`
}
