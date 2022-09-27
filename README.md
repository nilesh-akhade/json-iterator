# json-iterator
Low level iterator on the records inside large JSON file.

## Problem

How to read records from a very large and nested JSON file without loading the whole file in a memory?

If a JSON file is

- Very large
- The record to read is at deepest nested location
- May contain invalid json record

We cannot use `json.Marshal()` in this case, as it will load everything into memory causing out of memory error.

## Solution

Register the type decoder with the iterator and call `Next()` till `HasNext()` returns `false`.

    iterator := jsoniterator.NewJsonRecordsIterator("fruits.json")
    iterator.RegisterGoTypeDecoder("{.trees.[.{.fruits", func(d *json.Decoder, m map[string]interface{}) (interface{}, error) {
        var fruit Fruit
        err := d.Decode(&fruit)
        // enrich fruit with data from m
        return fruit, err
    })
    for iterator.HasNext() {
        item := iterator.Next()
        if fruit, ok := item.(Fruit); ok {
            fmt.Println(fruit)
        }
    }

See complete [example](cmd/fruits-example/main.go)
