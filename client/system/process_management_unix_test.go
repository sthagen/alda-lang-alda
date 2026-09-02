//go:build !windows

package system

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestSysProcAttr(t *testing.T) {
	attr := sysProcAttr()
	if attr == nil {
		t.Fatal("sysProcAttr() should return a non-nil value on non-Windows platforms")
	}

	if !attr.Setpgid {
		t.Errorf("sysProcAttr().Setpgid should be true, got: %v", attr.Setpgid)
	}
}
