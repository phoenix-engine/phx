package cpp

import (
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// Resource represents a static asset or resource generated from a file.
type Resource struct {
	// Name is the original filename of the resource.  This is
	// needed for encoding the actual variable name in the var
	// declaration.
	Name string

	// TODO:
	// ID is a numeric ID assigned to the resource in case of a name
	// collision.
	ID int

	// Size is the full uncompressed size of the resource.
	Size int64

	// Into is the writer which the static asset will be written to.
	// Decl is the writer which will encode the variable declaration
	// referring to the asset.
	Into, Decl io.WriteCloser
}

func (r *Resource) Write(some []byte) (n int, err error) {
	n, err = r.Into.Write(some)
	r.Size += int64(n)
	return
}

func (r *Resource) ReadFrom(some io.Reader) (n int64, err error) {
	if rf, ok := r.Into.(io.ReaderFrom); ok {
		n, err = rf.ReadFrom(some)
		r.Size += int64(n)
		return
	}

	n, err = io.Copy(r.Into, some)
	r.Size += int64(n)
	return
}

func (r *Resource) Close() (err error) {
	if err = r.Into.Close(); err != nil {
		return errors.Wrapf(err, "closing asset for %s", r.Name)
	}

	if err := Decl(*r).Expand(r.Decl); err != nil {
		return errors.Wrapf(err, "expanding declaration for %s", r.Name)
	}

	// Now close the declaration and output file.
	return errors.Wrapf(r.Decl.Close(), "closing declaration for %s", r.Name)
}

// VarName returns the cleansed name of the resource which may be used
// as a sanitized variable name in C++.
func (r Resource) VarName() string {
	firstOK := false
	return strings.Map(func(rr rune) rune {
		if !firstOK {
			firstOK = true
			if '0' <= rr && rr <= '9' {
				return 'd'
			}
		}

		switch {
		case 'A' <= rr && rr <= 'z',
			'0' <= rr && rr <= '9':
			return rr
		}

		// Special-cased matches
		switch rr {
		case '_', '-':
		}

		return '_'
	}, r.Name) // + r.ID
}

// Resources implements sort.Interface.
type Resources []Resource

// Resources implements sort.Interface.
func (r Resources) Len() int      { return len(r) }
func (r Resources) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Resources) Less(i, j int) bool {
	return r[i].VarName() < r[j].VarName()
}

type ResourceManager struct {
	Resources
}

type Decl Resource

func (d Decl) Expand(r io.WriteCloser) error {
	res := Resource(d)
	name := "res/" + res.VarName() + "_decl.cxx"

	tmp, err := template.New(name).Parse(templates[TmpDecl])
	if err != nil {
		return errors.Wrapf(err, "parsing %s template", name)
	}

	return errors.Wrapf(tmp.Execute(r, res), "executing %s", name)
}
