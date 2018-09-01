package cpp

import (
	"bytes"
	"io"
	"text/template"

	"github.com/synapse-garden/phx/fs"

	"github.com/pkg/errors"
)

type TemplateID int

// Template ID constants.
const (
	TmpDecl TemplateID = iota
	TmpID
	TmpMapperHdr
	TmpMapperImpl
	TmpMappings
	TmpResourceHdr
	TmpResourceImpl

	TmpCMakeLists
	TmpGitModules
	TmpGitignore
	TmpClangFormat
)

var templates = map[TemplateID]string{
	TmpDecl:         declTmp,
	TmpID:           idTmp,
	TmpMapperHdr:    mapperHdrTmp,
	TmpMapperImpl:   mapperImplTmp,
	TmpMappings:     mappingsTmp,
	TmpResourceHdr:  resourceHdrTmp,
	TmpResourceImpl: resourceImplTmp,

	TmpCMakeLists:  cmakeTmp,
	TmpGitModules:  gitModuleTmp,
	TmpGitignore:   gitignoreTmp,
	TmpClangFormat: clangFormatTmp,
}

// TODO: Tighten up this package a lot.
func Execute(f fs.FS, t *template.Template, args interface{}) error {
	ff, err := f.Create(t.Name())
	if err != nil {
		return errors.Wrapf(err, "creating %s", t.Name())
	}

	if err := t.Execute(ff, args); err != nil {
		return errors.Wrapf(err, "executing %s", t.Name())
	}

	return errors.Wrapf(ff.Close(), "closing %s", t.Name())
}

type Creator interface{ Create(fs.FS) error }

type ID Resources

// Create a header defining an enum of IDs.
func (i ID) Create(f fs.FS) error {
	return create(f, "id.hpp", TmpID, Resources(i))
}

type Mappings Resources

func (m Mappings) Create(f fs.FS) error {
	return create(f, "mappings.cxx", TmpMappings, Resources(m))
}

type MapperHdr Resources

func (m MapperHdr) Create(f fs.FS) error {
	return create(f, "mapper.hpp", TmpMapperHdr, Resources(m))
}

type CMakeLists Resources

func (c CMakeLists) Create(f fs.FS) error {
	return create(f, "CMakeLists.txt", TmpCMakeLists, Resources(c))
}

func create(f fs.FS, name string, id TemplateID, rs Resources) error {
	tmp, err := template.New(name).Parse(templates[id])
	if err != nil {
		return errors.Wrapf(err, "parsing %s template", name)
	}

	return Execute(f, tmp, rs)
}

// CreateImplementations creates the files that don't rely on variable
// state.
func CreateImplementations(f fs.FS) error {
	for fname, tmp := range map[string]TemplateID{
		"mapper.cxx":    TmpMapperImpl,
		"resource.hpp":  TmpResourceHdr,
		"resource.cxx":  TmpResourceImpl,
		".gitmodules":   TmpGitModules,
		".gitignore":    TmpGitignore,
		".clang-format": TmpClangFormat,
	} {
		ff, err := f.Create(fname)
		if err != nil {
			return errors.Wrapf(err, "creating %s", fname)
		}

		_, err = io.Copy(ff, bytes.NewBufferString(templates[tmp]))
		if err != nil {
			return errors.Wrapf(err, "writing %s", fname)
		}

		if err := ff.Close(); err != nil {
			return errors.Wrapf(err, "closing %s", fname)
		}
	}

	return nil
}
