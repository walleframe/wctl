package parser

import (
	"encoding/json"
	"testing"
)

func TestCheck(t *testing.T) {
	RegisterCheck("range", CheckIntegerRange)
	list, err := ParseCheker("range(1,2)")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("size:", len(list))
	list2, err := ParseCheker("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("size", len(list2))

	v := TTT{
		V: map[int]int{1: 2},
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

type TTT struct {
	V map[int]int
}
