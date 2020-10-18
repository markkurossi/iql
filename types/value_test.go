//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package types

import (
	"testing"
)

func TestBool(t *testing.T) {
	val := BoolValue(false)

	_, err := val.Bool()
	if err != nil {
		t.Errorf("Bool() failed: %s", err)
	}
	str := val.String()
	if str != False {
		t.Errorf("Bool.String() failed: got %s, expected %s", str, False)
	}
}

func TestInt(t *testing.T) {
	val := IntValue(0)

	_, err := val.Int()
	if err != nil {
		t.Errorf("Int() failed: %s", err)
	}
	_, err = val.Float()
	if err != nil {
		t.Errorf("Float() failed: %s", err)
	}
}

func TestFloat(t *testing.T) {
	val := FloatValue(0.0)

	_, err := val.Int()
	if err != nil {
		t.Errorf("Int() failed: %s", err)
	}
	_, err = val.Float()
	if err != nil {
		t.Errorf("Float() failed: %s", err)
	}
}
