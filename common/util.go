package common

import (
	"bytes"
	"encoding/gob"
)

func GobEncode(v any) []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes()
}

func GobDecode(data []byte, out any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(out)
}
