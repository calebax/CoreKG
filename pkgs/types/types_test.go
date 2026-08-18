package types

import (
	"encoding/json"
	"testing"
)

func TestSafeID(t *testing.T) {
	type SA struct {
		UserID SafeID `json:"user_id"`
		Name   Secret `json:"name"`
	}
	sa := SA{
		UserID: 123456,
		Name:   "123456sdaf",
	}
	b, err := json.Marshal(sa)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
	var sa2 SA
	err = json.Unmarshal(b, &sa2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(sa2)

}
