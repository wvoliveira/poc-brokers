package main

import (
	"encoding/json"
	"fmt"
)

func EncodeFind(payload []byte) (r Response, err error) {
	fmt.Printf("Encode payload: %s\n", string(payload))
	err = json.Unmarshal([]byte(payload), &r)
	return
}
