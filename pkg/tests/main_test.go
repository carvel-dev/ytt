package tests

import (
	"testing"
)

func TestPkg2(t *testing.T) {
	pkg := &Package{
		Location: "pkg2",
	}
	pkg.Test(t)
}
