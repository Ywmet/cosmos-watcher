package block

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// This function is used only to dump blocks to files in order to see if we are receiving them correctly
func TestBlockStream(t *testing.T) {
	stream, err := GetBlockStream(context.Background(), "tcp://localhost:26657")
	if err != nil {
		t.Fatal(err)
	}

	for block := range stream {
		data, err := json.MarshalIndent(block, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(data))
		fmt.Println("---------------------------------------------------------------------------------------------------------------------------")
	}
}
