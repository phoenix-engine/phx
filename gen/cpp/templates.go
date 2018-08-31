package cpp

// TODO: Read templates from a template directory.
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
{{end}}#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
{{range .}}	{{template "expand" .}}{{end}}    };
};

#endif
`[1:]

var mappingTmp = `
#ifndef PHX_RES_ID
#define PHX_RES_ID

namespace res {
    enum class ID {
	{{range ids}}{{{end}}
    };
};

#endif
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
