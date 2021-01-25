//
// Copyright (c) 2021 Markku Rossi
//
// All rights reserved.
//

package iql

import (
	"os"
	"testing"

	"github.com/markkurossi/iql/lang"
)

func TestClient(t *testing.T) {
	client := NewClient(os.Stdout)
	err := client.SetString(lang.SysTableFmt, "ascii")
	if err != nil {
		t.Errorf("client.SetString(SysTableFmt): %s", err)
	}
}
