package encoding

import (
	"bytes"
	"encoding/gob"
)

// Marshal takes an interface and encodes it in a bytes slice
func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Unmarshal takes encoded bytes slice and decodes it in the passed interface
func Unmarshal(data []byte, v interface{}) (err error) {
	b := bytes.NewBuffer(data)
	err = gob.NewDecoder(b).Decode(v)
	return
}
