package cpp_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/phoenix-engine/phx/fs"
	"github.com/phoenix-engine/phx/gen/cpp"
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

func TestCMakeCreator(t *testing.T) {
	expect := `
cmake_minimum_required(VERSION 3.1.0 FATAL_ERROR)

# Add LZ4 and LZ4F definitions.
add_subdirectory(lz4/lib)

# Add Resource library.
add_library(Resource STATIC
  mapper.cxx
  mappings.cxx
  resource.cxx
  res/al_gif_decl.cxx
  res/al_jpg_decl.cxx
  res/bob_gif_decl.cxx
  res/bob_jpg_decl.cxx
)

target_include_directories(Resource PUBLIC
  ${CMAKE_CURRENT_LIST_DIR}
)

target_link_libraries(Resource LZ4F)

set_property(TARGET Resource PROPERTY CXX_STANDARD 11)
set_property(TARGET Resource PROPERTY CXX_STANDARD_REQUIRED ON)
`[1:]

	ff := mockFS{objs: make(map[string]bcl)}
	ii := cpp.CMakeLists{
		{Name: "al.gif"},
		{Name: "al.jpg"},
		{Name: "bob.gif"},
		{Name: "bob.jpg"},
	}

	if err := ii.Create(ff); err != nil {
		t.Errorf("expected nil error, got %#v", err)
		t.FailNow()
	}

	t.Log("mock FS now contains CMakeLists.txt with the defined IDs")

	if len(ff.objs) != 1 {
		t.Errorf("unexpected objects in FS: %+v", ff.objs)
	}
	if fs := ff.objs["CMakeLists.txt"].String(); fs != expect {
		t.Errorf("\n======== expected:\n%s\n\n"+
			"======== got:\n%s", expect, fs)
		t.FailNow()
	}
}

func TestMappingsCreator(t *testing.T) {
	expect := `
#include "id.hpp"
#include "mapper.hpp"

namespace res {
    std::map<ID, const Mapper::resDefn> Mapper::mappings{

	// al.gif
	{
		ID::al_gif,
		{ 100, al_gif_len, al_gif },
	},

	// al.jpg
	{
		ID::al_jpg,
		{ 200, al_jpg_len, al_jpg },
	},

	// bob.gif
	{
		ID::bob_gif,
		{ 300, bob_gif_len, bob_gif },
	},

	// bob.jpg
	{
		ID::bob_jpg,
		{ 400, bob_jpg_len, bob_jpg },
	},
    };
}; // namespace res
`[1:]

	ff := mockFS{objs: make(map[string]bcl)}
	ii := cpp.Mappings{
		{Name: "al.gif", CompCount: 100},
		{Name: "al.jpg", CompCount: 200},
		{Name: "bob.gif", CompCount: 300},
		{Name: "bob.jpg", CompCount: 400},
	}

	if err := ii.Create(ff); err != nil {
		t.Errorf("expected nil error, got %#v", err)
		t.FailNow()
	}

	t.Log("mock FS now contains mappings.cxx with the defined IDs")

	if len(ff.objs) != 1 {
		t.Errorf("unexpected objects in FS: %+v", ff.objs)
	}
	if fs := ff.objs["mappings.cxx"].String(); fs != expect {
		t.Errorf("\n======== expected:\n%s"+
			"======== got:\n%s", expect, fs)

		var line string
		linenum := 1

		for i := range expect {
			if expect[i] != fs[i] {
				t.Logf("fs[%d](%#q) didn't match expect[%d](%#q).",
					i, fs[i], i, expect[i],
				)
				t.Logf("fs line %d: %#q...", linenum, line)
				break
			}

			if fs[i] == '\n' {
				linenum++
				line = ""
			} else {
				line += string(fs[i])
			}
		}
		t.FailNow()
	}
}

func TestMapperCreator(t *testing.T) {
	expect := `
#ifndef PHX_RES_MAPPER
#define PHX_RES_MAPPER

#include <memory>
#include <map>

#include "id.hpp"
#include "resource.hpp"

namespace res {
    // Mapper encapsulates implementation details of the mapping of IDs
    // to Resources away from the user.
    //
    // Fetch is used to retrieve a new Resource, which can be used to
    // decompress a static asset.  It does not create a new copy of the
    // asset.
    class Mapper {
    public:
	// Mapper may not be instantiated.
	Mapper() = delete;

	// Fetch creates and retrieves a unique smart-pointer to a
	// Resource.
	static std::unique_ptr<Resource> Fetch(ID) noexcept(false);

    private:
	struct resDefn {
	    size_t               compressed_length;
	    size_t               decompressed_length;
	    const unsigned char* content;
	};

	static std::map<ID, const resDefn> mappings;

	// Here, all names of assets are defined.  Each must have an ID
	// associated with it.

	// al.gif
	static const size_t        al_gif_len;
	static const unsigned char al_gif[];

	// al.jpg
	static const size_t        al_jpg_len;
	static const unsigned char al_jpg[];

	// bob.gif
	static const size_t        bob_gif_len;
	static const unsigned char bob_gif[];

	// bob.jpg
	static const size_t        bob_jpg_len;
	static const unsigned char bob_jpg[];
    };
}; // namespace res

#endif
`[1:]

	ff := mockFS{objs: make(map[string]bcl)}
	ii := cpp.MapperHdr{
		{Name: "al.gif"},
		{Name: "al.jpg"},
		{Name: "bob.gif"},
		{Name: "bob.jpg"},
	}

	if err := ii.Create(ff); err != nil {
		t.Errorf("expected nil error, got %#v", err)
		t.FailNow()
	}

	t.Log("mock FS now contains mapper.hpp with the defined IDs")

	if len(ff.objs) != 1 {
		t.Errorf("unexpected objects in FS: %+v", ff.objs)
	}
	if fs := ff.objs["mapper.hpp"].String(); fs != expect {
		t.Errorf("\n======== expected:\n%s\n\n"+
			"======== got:\n%s", expect, fs)
		t.FailNow()
	}
}
