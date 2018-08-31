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
	Split(path string) (parent, rest string)
}

// DirMaker is an FS which has a meaningful concept of a directory.
// This prevents objects being moved into a directory which does not
// exist, so the directory must be created.
//
// Mkdir should create any necessary parent directories with the same
// privilege passed as "perm".
type DirMaker interface {
	Mkdir(perm os.FileMode, path ...string) error
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

func (r Real) Split(path string) (parent, rest string) {
	return filepath.Split(path)
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

func (r Real) Mkdir(perm os.FileMode, path ...string) error {
	return os.MkdirAll(
		r.Join(append([]string{r.Where}, path...)...),
		perm,
	)
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

		fromReal, toReal := from.(Real), to.(Real)

		// Moving from one prefix to another.
		err := os.Rename(
			filepath.Join(fromReal.Where, pFrom),
			filepath.Join(toReal.Where, pTo),
		)
		if err == nil {
			return nil
		}

		if os.IsNotExist(err) {
			// The target directory didn't exist.
			parentTo, _ := toReal.Split(pTo)
			err = toReal.Mkdir(0755, parentTo)
			if err != nil {
				return errors.Wrapf(err,
					"creating %s",
					parentTo)
			}
		}
		return os.Rename(
			filepath.Join(fromReal.Where, pFrom),
			filepath.Join(toReal.Where, pTo),
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
	switch {
	case err == nil:
	case os.IsNotExist(err):
		// Target directory needs to be created.
		if toMk, ok := to.(DirMaker); ok {
			parentTo, _ := to.Split(pTo)
			err := toMk.Mkdir(0755, parentTo)
			if err != nil {
				return errors.Wrapf(err,
					"creating %s",
					parentTo)
			}
			if bTo, err = to.Create(pTo); err != nil {
				return errors.Wrapf(err,
					"creating destination %s",
					pTo,
				)
			}
			break
		}

		fallthrough

	case err != nil:
		// Some other problem happened.
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
