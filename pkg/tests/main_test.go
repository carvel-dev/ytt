package tests

import (
	"testing"
)

func TestPkg1(t *testing.T) {
	pkg := &Package{
		Location:  "pkg1",
		Recursive: true,
	}
	pkg.Test(t)
}
