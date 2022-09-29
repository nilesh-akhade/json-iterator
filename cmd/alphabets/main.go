package main

import (
	"encoding/json"
	"fmt"

	jsoniterator "github.com/nilesh-akhade/json-iterator"
)

func main() {
	iterator := jsoniterator.NewJsonRecordsIterator("alphabets.json")
	iterator.RegisterTypeDecoder("{}{.a.{.b.{.c.{.d", func(d *json.Decoder, m map[string]interface{}) (interface{}, error) {
		var bag BagOfAlpha
		err := d.Decode(&bag)
		if val, found := m["{.a.{.b.{.level"]; found {
			if bLevel, ok := val.(float64); ok {
				bag.BLevel = bLevel
			}
		}
		return bag, err
	})
	for iterator.HasNext() {
		item := iterator.Next()
		if bag, ok := item.(BagOfAlpha); ok {
			fmt.Printf("%#v\n", bag)
		}
	}
	if iterator.Err() != nil {
		fmt.Println(iterator.Err())
	}
}

type BagOfAlpha struct {
	A      string
	B      string
	BLevel float64
	C      string
	D      string
	E      string
}
