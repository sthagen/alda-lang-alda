//go:build windows

package system

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestSysProcAttr(t *testing.T) {
	attr := sysProcAttr()
	if attr == nil {
		t.Fatal("sysProcAttr() should not return nil on Windows — " +
			"a nil SysProcAttr means the player process would remain " +
			"attached to the parent console and be killed when the " +
			"terminal closes (see issue #406)")
	}
	if attr.CreationFlags&detachedProcess == 0 {
		t.Errorf(
			"sysProcAttr().CreationFlags should include detachedProcess "+
				"(0x%x) so the player survives the terminal closing, got: 0x%x",
			detachedProcess, attr.CreationFlags,
		)
	}
}
