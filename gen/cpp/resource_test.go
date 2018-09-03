package cpp_test

import (
	"bytes"
	"testing"

	"github.com/phoenix-engine/phx/gen/cpp"
)

func TestDeclExpand(t *testing.T) {
	expect := `
#include "mapper.hpp"

namespace res {
    const size_t        Mapper::foo_txt_len = 1000;
    const unsigned char Mapper::foo_txt[]   = {
#include "foo_txt_real.cxx"
    };
}; // namespace res
`[1:]

	buf := bcl{new(bytes.Buffer)}

	err := cpp.AssetDecl(cpp.Resource{Name: "foo.txt", Size: 1000}).Expand(buf)

	if err != nil {
		t.Errorf("expected nil error, but got %#v", err)
		t.FailNow()
	}

	if bs := buf.String(); bs != expect {
		t.Errorf("\n======== expected:\n%s\n\n"+
			"======== got:\n%s", expect, bs)
		t.FailNow()
	}
}
