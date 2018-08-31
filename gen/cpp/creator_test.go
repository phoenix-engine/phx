package cpp_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/synapse-garden/phx/fs"
	"github.com/synapse-garden/phx/gen/cpp"
)

type bcl struct{ *bytes.Buffer }

func (bcl) Close() error { return nil }

type mockFS struct {
	fs.FS
	objs map[string]bcl
}

func (m mockFS) Create(name string) (io.WriteCloser, error) {
	buf := bcl{new(bytes.Buffer)}
	m.objs[name] = buf
	return buf, nil
}

func TestIDCreator(t *testing.T) {
	expect := `
#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
	al_gif, // al.gif
	al_jpg, // al.jpg
	bob_gif, // bob.gif
	bob_jpg, // bob.jpg
    };
};

#endif
`[1:]

	ff := mockFS{objs: make(map[string]bcl)}
	ii := cpp.ID{
		{Name: "al.gif"},
		{Name: "al.jpg"},
		{Name: "bob.gif"},
		{Name: "bob.jpg"},
	}

	if err := ii.Create(ff); err != nil {
		t.Errorf("expected nil error, got %#v", err)
		t.FailNow()
	}

	t.Log("mock FS now contains id.hpp with 4 IDs defined in enum")

	if len(ff.objs) != 1 {
		t.Errorf("unexpected objects in FS: %+v", ff.objs)
	}
	if fs := ff.objs["id.hpp"].String(); fs != expect {
		t.Errorf("\n======== expected:\n%s\n\n"+
			"======== got:\n%s", expect, fs)
		t.FailNow()
	}
}
