package model

import (
	"testing"
)

func TestLease(t *testing.T) {
	l := &Lease{
		// PublicKey: "a",
		// Contract:  "b",
		// Expire:    "c",
		Resources: Resources{
			Cpu: "d",
		},
	}
	b, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}

	la := &Lease{}
	err = la.Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(la)
}
