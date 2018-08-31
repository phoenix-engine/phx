package cpp

import (
	"text/template"

	"github.com/pkg/errors"
	"github.com/synapse-garden/phx/fs"
)

type TemplateID int

// Template ID constants.
const (
	IDNames TemplateID = iota
	IDID
	IDMapping
	IDMapper
)

var templates = map[TemplateID]string{
	IDNames:   namesTmp,
	IDID:      idTmp,
	IDMapping: mappingTmp,
	IDMapper:  mapperTmp,
}

func Execute(f fs.FS, t *template.Template, args interface{}) error {
	ff, err := f.Create("id.hpp")
	if err != nil {
		return errors.Wrapf(err, "creating %s", t.Name())
	}

	if err := t.Execute(ff, args); err != nil {
		return errors.Wrapf(err, "executing %s", t.Name())
	}

	return errors.Wrapf(ff.Close(), "closing %s", t.Name())
}

type Creator interface {
	Create(fs.FS) error
}

type Names Resources

func (n Names) Create(f fs.FS) error {
	// Create a list of names. (???)
	return nil
}

type ID Resources

// Create a header defining an enum of IDs.
func (i ID) Create(f fs.FS) error {
	tmp, err := template.New("id.hpp").Parse(templates[IDID])
	if err != nil {
		return errors.Wrap(err, "parsing id.hpp template")
	}

	return Execute(f, tmp, i)
}

type Mapping Resources

func (m Mapping) Create(f fs.FS) error {
	return nil
}

type Mapper Resources

func (m Mapper) Create(f fs.FS) error {
	return nil
}
