package cpp

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/synapse-garden/phx/fs"
	"github.com/synapse-garden/phx/gen/compress"
)

// Size constants.
const (
	PageSize = 4096
	MaxWidth = 72
)

type CloserCloser struct {
	first, second io.Closer
}

func (c CloserCloser) Close() error {
	if err := c.first.Close(); err != nil {
		return err
	}
	return c.second.Close()
}

func PrepareTarget(over fs.FS, using compress.Maker) Target {
	return Target{
		FS:        over,
		WaitGroup: new(sync.WaitGroup),
		// The Target has a Pool of compressors, which will be
		// created and returned as needed.
		Pool: &sync.Pool{
			New: func() interface{} { return using.Make() },
		},

		done: make(chan Resource),
	}
}

// Target is a complete C++ static asset class.
//
// TODO: cpp.Target is just a wrapper for a handful of C helpers.
type Target struct {
	fs.FS
	*sync.WaitGroup
	*sync.Pool

	done chan Resource

	// TODO: Cancel()
	cancel chan struct{}
}

// Create creates a Resource which the static asset will be written to,
// which uses a Compressor from the Target's pool.
func (t Target) Create(name string) (io.WriteCloser, error) {

	// Create a Resource to manage the creation of the asset and its
	// variable declaration.  The project layout is created in
	// Finalize() using the full Resource list.
	res := &Resource{Name: name}

	// Create the asset container (e.g. "dat_txt_real.cxx".)
	assetF, err := t.FS.Create(t.FS.Join("res", res.VarName()+"_real.cxx"))
	if err != nil {
		return nil, errors.Wrapf(err, "creating asset %s", name)
	}

	// Create the variable declaration file for the resource (e.g.
	// "dat_txt_decl.cxx".)
	declF, err := t.FS.Create(t.FS.Join("res", res.VarName()+"_decl.cxx"))
	if err != nil {
		return nil, errors.Wrapf(err, "creating decl %s", name)
	}

	// Fetch and prepare a new Compressor over an ArrayWriter.
	comp := t.Get().(compress.Compressor)
	aw := NewArrayWriter(assetF)
	comp.Reset(aw)

	// The raw asset will be copied into "Into", which compresses
	// and writes the compressed content into the ArrayWriter, which
	// encodes the compressed bytes as a C++ array literal.
	//
	// The Compressor may also implement compress.Counter, counting
	// the compressed bytes.
	res.Into, res.Decl = comp, declF

	// This takes care of flushing the compressor and array writer
	// first, and then closing the underlying buffer or file.
	res.CloserCloser = CloserCloser{first: res.Into, second: assetF}

	done := make(chan struct{})
	t.Add(1)

	go func() {
		// Decrement the waitgroup when all finished.
		defer t.Done()

		// Wait for the Resource to be Closed.  When it is, it
		// wil be added to the list of finished resources.
		<-done

		// Reset and recycle the Compressor.
		comp.Reset(nil)
		t.Put(comp)

		// t.done is unbuffered, so every Resource will have a
		// waiting channel send after it's finished encoding.
		// These will be consumed in Finalize().
		t.done <- *res
	}()

	return DoneCloser{res, done}, nil
}

type allErrs []error

func (a allErrs) Error() string {
	var b []string
	for _, e := range a {
		b = append(b, e.Error())
	}
	return fmt.Sprintf("%d errors: "+strings.Join(b, "; "), len(b))
}

func (t Target) Finalize() error {
	var res Resources
	go func() {
		for re := range t.done {
			// Each one represents two files.
			res = append(res, re)
		}
	}()

	// Create all the files which don't rely on variable state.
	// They can be created while the resources are still being
	// processed, but we'll wait at least until Finalize is called.
	if err := CreateImplementations(t.FS); err != nil {
		t.Wait()
		close(t.done)
		return errors.Wrap(err, "creating implementation files")
	}

	// Wait for all Resource names to be processed so we can use
	// them in the Mapper, etc.
	t.Wait()
	close(t.done)

	sort.Sort(res)

	// TODO: use templates from a subrepo / subfolder.
	ccs := []Creator{
		ID(res),
		Mappings(res),
		MapperHdr(res),
		CMakeLists(res),
	}

	errs := make(chan error)
	for _, cc := range ccs {
		go func(c Creator) { errs <- c.Create(t.FS) }(cc)
	}

	var ees allErrs
	for i := 0; i < len(ccs); i++ {
		if err := <-errs; err != nil {
			ees = append(ees, err)
		}
	}

	if ees == nil {
		// A nil slice isn't a nil interface.  If we return ees,
		// the error interface will be non-nil, containing a nil
		// slice as its implementation.
		return nil
	}
	return ees
}
