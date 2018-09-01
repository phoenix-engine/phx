package cpp

// TODO: Read templates from a template directory.
// TODO: Add a "phx gen" helper for cloning / browsing templates

var declTmp = `
#include "mapper.hpp"

namespace res {
    const size_t        Mapper::{{.VarName}}_len = {{.Size}};
    const unsigned char Mapper::{{.VarName}}[]   = {
#include "{{.VarName}}_real.cxx"
    };
}; // namespace res
`[1:]

var idTmp = `
{{define "expand"}}{{.VarName}}, // {{.Name}}
{{end}}`[1:] + `
#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
{{range .}}	{{template "expand" .}}{{end}}    };
};

#endif
`[1:]

var mappingsTmp = `
{{define "expand"}}
	// res/{{.Name}}
	{ ID::{{.VarName}},
	  {
	    .compressed_length   = std::extent<decltype({{.VarName}})>::value,
	    .decompressed_length = {{.VarName}}_len,
	    .content             = {{.VarName}},
	  } },{{end}}`[1:] + `

#include "id.hpp"
#include "mapper.hpp"

namespace res {
    std::map<ID, const Mapper::resDefn> Mapper::mappings{`[2:] + `

{{range .}}{{template "expand" .}}
{{end}}
`[1:] + `
    };
}; // namespace res
`[1:]

var mapperHdrTmp = `
{{define "expand"}}
	// {{.Name}}
	static const size_t        {{.VarName}}_len;
	static const unsigned char {{.VarName}}[];{{end}}`[1:] + `

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
`[2:] + `

{{range .}}{{template "expand" .}}
{{end}}`[2:] + `

    };
}; // namespace res

#endif
`[2:]

var mapperImplTmp = `
#include <map>

#include "lz4frame.h"

#include "id.hpp"
#include "mapper.hpp"
#include "resource.hpp"

namespace res {
    std::unique_ptr<Resource> Mapper::Fetch(ID id) noexcept(false) {
	// This will never fail as long as every ID has a mapping.
	auto from = mappings[id];

	// TODO: Decouple context creation from Mapper.
	LZ4F_dctx* dec;
	auto       lz4v = LZ4F_getVersion();
	auto       err  = LZ4F_createDecompressionContext(&dec, lz4v);

	if (LZ4F_isError(err)) {
	    throw LZ4F_getErrorName(err);
	}

	return std::move(std::unique_ptr<Resource>(
	  new Resource(dec, from.content, from.compressed_length,
	               from.decompressed_length)));
    };
}; // namespace res
`[1:]

var resourceImplTmp = `
#include "lz4frame.h"

#include "resource.hpp"

namespace res {
    namespace {
	LZ4F_frameInfo_t _noFrame;

	using bid = LZ4F_blockSizeID_t;
	const size_t lookupBlkSize(bid szid) noexcept(true) {
	    switch (szid) {
	    case LZ4F_default:
		return 2 << 21;
	    case LZ4F_max64KB:
		return 2 << 15;
	    case LZ4F_max256KB:
		return 2 << 17;
	    case LZ4F_max1MB:
		return 2 << 19;
	    case LZ4F_max4MB:
		return 2 << 21;
	    default:
		return 2 << 21;
	    }
	}
    }; // namespace

    size_t Resource::Len() noexcept(true) {
	return decompressed_content_length;
    }

    Resource::Resource(
      LZ4F_dctx* decoder, const unsigned char* content,
      size_t compressed_content_length,
      size_t decompressed_content_length) noexcept(true)
        : decoder(decoder), consumed(0), next_read_size(0),
          compressed_content_length(compressed_content_length),
          decompressed_content_length(decompressed_content_length),
          content(content) {}

    Resource::~Resource() noexcept(false) {
	// TODO: Instead of throwing an exception, free the decoder in
	// some other way, such as when consumed == length of content.
	auto err = LZ4F_freeDecompressionContext(decoder);
	if (LZ4F_isError(err)) {
	    throw LZ4F_getErrorName(err);
	}
    }

    const size_t Resource::BlockSize() noexcept(false) {
	LZ4F_frameInfo_t frame = _noFrame;

	// "more" is how much max will be parsed from buf as the header.
	size_t more = LZ4F_HEADER_SIZE_MAX;

	auto errOrNext = LZ4F_getFrameInfo(decoder, &frame,
	                                   content + consumed, &more);
	if (LZ4F_isError(errOrNext)) {
	    throw LZ4F_getErrorName(errOrNext);
	}

	// "more" is now how much was read from buf.
	consumed += more;

	// "errOrNext" is the expected size of the next read.
	next_read_size = errOrNext;

	return lookupBlkSize(frame.blockSizeID);
    }

    const size_t Resource::Read(char*  into,
                                size_t len) noexcept(false) {
	size_t intoSize = len;
	size_t more     = next_read_size;
	size_t written  = 0;
	bool   done     = false;

	// Precalculate the size of the next read.
	BlockSize();

	while (!done && more > 0 && written < len) {
	    // Consume the frame into the destination.
	    auto errOrMore = LZ4F_decompress(
	      decoder, into + written, &intoSize, content + consumed,
	      &more, NULL);
	    if (LZ4F_isError(errOrMore)) {
		throw LZ4F_getErrorName(errOrMore);
	    }
	    if (errOrMore == 0) {
		// TODO: Maybe there's a better way to do this?
		done = true;
	    }

	    // "intoSize" is now the amount actually decoded into
	    // "into", so add it to the total written.
	    written += intoSize;

	    // Reset "intoSize" to the remaining target buffer for the
	    // next call.
	    intoSize = len - written;

	    // After decoding the block, "more" is the count of bytes
	    // consumed, in that call, from the internal compressed
	    // buffer source.  Add this to the total internal offset.
	    consumed += more;

	    // errOrMore is the ideal next read size.
	    more = next_read_size = errOrMore;
	}

	return written;
    }

    void Resource::Reset() noexcept(true) {
	LZ4F_resetDecompressionContext(decoder);

	consumed = next_read_size = 0;
    }
}; // namespace res
`[1:]

var resourceHdrTmp = `
#ifndef PHX_RES
#define PHX_RES

#include "lz4frame.h"

namespace res {
    class Resource {
    public:
	// The default constructor is meaningless.  Every Resource must
	// be created with a reference to a static array with an
	// uncompressed length.
	Resource() = delete;

	// To construct a Resource, pass it an initialized LZ4F decoder
	// context, an array containing LZ4 compressed bytes, and the
	// size (in bytes) of the uncompressed resource.
	//
	// Most users should simply use Mapper::Fetch.
	Resource(LZ4F_dctx*, const unsigned char*,
	         size_t compressed_length,
	         size_t decompressed_length) noexcept(true);
	~Resource() noexcept(false);

	// Len returns the full decompressed size of the asset.
	size_t Len() noexcept(true);

	// BlockSize returns the maximum required size of a block which
	// may be written to by a single partial Read.
	//
	// TODO: Make this a feature of a subclass or something so the
	// streaming mode doesn't clutter the simple API.
	const size_t BlockSize() noexcept(false);

	// Read ingests up to len bytes into the target buffer.  For the
	// best performance, the user should pass a buffer sized to the
	// full size of the resource, given by Len(), or to the block
	// size, given by BlockSize().  A smaller buffer may also be
	// used.
	//
	// Reset() may be called to begin from the beginning.
	const size_t Read(char* into, size_t len) noexcept(false);

	// Reset returns the state of the Res to its initial state,
	// ready to begin filling a new target buffer.
	void Reset() noexcept(true);

    private:
	LZ4F_dctx* decoder;
	size_t     consumed;
	size_t     next_read_size;

	const size_t         compressed_content_length;
	const size_t         decompressed_content_length;
	const unsigned char* content;
    };
}; // namespace res

#endif
`[1:]

var cmakeTmp = `
cmake_minimum_required(VERSION 3.1.0 FATAL_ERROR)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Add LZ4 and LZ4F definitions.
add_subdirectory(lz4/lib)

# Add Resource library.
add_library(Resource STATIC
  mapper.cxx
  mappings.cxx
  resource.cxx
  res/dat_txt.cxx
)

target_include_directories(Resource PUBLIC
  ${CMAKE_CURRENT_LIST_DIR}
  ${LZ4F_INCLUDE_DIR}
)
target_link_libraries(Resource LZ4F)

set_property(TARGET Resource PROPERTY CXX_STANDARD 11)
set_property(TARGET Resource PROPERTY CXX_STANDARD_REQUIRED ON)

add_executable(ResTest main.cxx)
target_link_libraries(ResTest Resource)

set_property(TARGET ResTest PROPERTY CXX_STANDARD 11)
set_property(TARGET ResTest PROPERTY CXX_STANDARD_REQUIRED ON)
`[1:]

var gitModuleTmp = `
[submodule "lz4"]
	path = lz4
	url = git@github.com:synapse-garden/lz4.git
`[1:]

var gitignoreTmp = `
cmake_install.cmake
CMakeCache.txt
CMakeFiles/
build/*
!build/.gitkeep
lib/
!src/lib
target/

*.sw[nop]
*~
\#*
*#

*.so
*.a

*.opensdf
*.sdf
*.sln
*.suo
*.vcxproj
*.vcxproj.filters
Debug/
obj/
*.psess
*.vspx

.DS_Store
`[1:]

var clangFormatTmp = `
BasedOnStyle:                 LLVM
AlignConsecutiveAssignments:  true
AlignConsecutiveDeclarations: true
AccessModifierOffset:         -4
BinPackArguments:             true
BinPackParameters:            true
BreakStringLiterals:          true
ColumnLimit:                  72
ContinuationIndentWidth:      2
Cpp11BracedListStyle:         false
IndentCaseLabels:             false
IndentWidth:                  4
Language:                     Cpp
NamespaceIndentation:         All
PenaltyBreakAssignment:       100
PointerAlignment:             Left
ReflowComments:               true
SortIncludes:                 true
SpacesInContainerLiterals:    true
Standard:                     Auto
UseTab:                       ForIndentation
`[1:]
