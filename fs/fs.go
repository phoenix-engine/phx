package fs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	kfs "github.com/kr/fs"
	"github.com/pkg/errors"
)

type FS interface {
	kfs.FileSystem

	Open(string) (io.ReadCloser, error)
	Create(string) (io.WriteCloser, error)
	Move(from, to string) error
}

type Kind int

const (
	KindReal Kind = iota
	KindMem
)

var kinds = map[reflect.Type]Kind{
	reflect.TypeOf(Real{}): KindReal,
	reflect.TypeOf(Mem{}):  KindMem,
}

func KindOf(of reflect.Type) Kind {
	return kinds[of]
}

// Move is not concurrency-safe.
func Move(from, to FS, pFrom, pTo string) error {
	tFrom, tTo := reflect.TypeOf(from), reflect.TypeOf(to)
	kf := KindOf(tFrom)
	switch {
	case tFrom != tTo:
		// We don't have the same type, so we can't move.

	case kf == KindReal:
		if from == to {
			// Moving within the same prefix.
			return from.Move(pFrom, pTo)
		}

		// Moving from one prefix to another.
		return os.Rename(
			filepath.Join(from.(Real).Where, pFrom),
			filepath.Join(to.(Real).Where, pTo),
		)

	case kf == KindMem:
		if from == to {
			// Can only move within the same fs.Mem.
			return from.(Mem).Move(pFrom, pTo)
		}

		to.(Mem).bufs[pTo] = from.(Mem).bufs[pFrom]
		delete(from.(Mem).bufs, pFrom)

		return nil
	}

	// Otherwise, copy.
	bFrom, err := from.Open(pFrom)
	if err != nil {
		return errors.Wrapf(err, "opening source %s", pFrom)
	}

	bTo, err := to.Create(pTo)
	if err != nil {
		return errors.Wrapf(err, "creating destination %s", pTo)
	}

	if _, err := io.Copy(bTo, bFrom); err != nil {
		return errors.Wrapf(err,
			"copying source %s to destination %s",
			pFrom, pTo,
		)
	}
	if err := bFrom.Close(); err != nil {
		return errors.Wrapf(err, "closing source %s", pFrom)
	}
	if err := bTo.Close(); err != nil {
		return errors.Wrapf(err, "closing destination %s", pTo)
	}

	return nil
}

type Real struct{ Where string }

func (r Real) ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(r.Join(r.Where, name))
}

func (r Real) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(r.Join(r.Where, path))
}

func (r Real) Join(ps ...string) string {
	return filepath.Join(ps...)
}

func (r Real) Open(name string) (io.ReadCloser, error) {
	return os.Open(r.Join(r.Where, name))
}

func (r Real) Create(name string) (io.WriteCloser, error) {
	return os.Create(r.Join(r.Where, name))
}

func (r Real) Move(from, to string) error {
	return os.Rename(
		r.Join(r.Where, from),
		r.Join(r.Where, to),
	)
}
