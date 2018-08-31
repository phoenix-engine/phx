package cpp

// TODO: Read templates from a template directory.
// TODO: Add a "phx gen" helper for cloning / browsing templates
var namesTmp = `
#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
	dat_txt, // dat.txt
    };
};

#endif
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

var mapperTmp = `
#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
	dat_txt, // dat.txt
    };
};

#endif
`[1:]
