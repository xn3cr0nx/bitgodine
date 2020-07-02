package encoding

import (
	"bytes"
	"encoding/gob"

	"github.com/vmihailenco/msgpack"
)

// GobMarshal takes an interface and encodes it in a bytes slice with gob
func GobMarshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GobUnmarshal takes encoded bytes slice and decodes it in the passed interface with gob
func GobUnmarshal(data []byte, v interface{}) (err error) {
	b := bytes.NewBuffer(data)
	err = gob.NewDecoder(b).Decode(v)
	return
}

// Marshal takes an interface and encodes it in a bytes slice
func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal takes encoded bytes slice and decodes it in the passed interface
func Unmarshal(data []byte, v interface{}) (err error) {
	return msgpack.Unmarshal(data, v)
}
