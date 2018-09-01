package cpp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/pkg/errors"
	"github.com/synapse-garden/phx/fs"
	"github.com/synapse-garden/phx/gen/compress"
)

// Size constants.
const (
	PageSize = 4096
	MaxWidth = 72
)

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
	res.Into, res.Decl = aw, declF

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

// ArrayWriter consumes bytes and formats them as a C++ array literal.
//
// TODO: Set an unrecoverable error state to indicate that internal
// state may be unusable.
//
// TODO: Support chunked write / partial lines from Write, etc. without
// weird line breaks.
//
// TODO: Support Windows-style newline + carriage return?
//
// TODO: Optimize for space?  (Base64 / other solutions?)
//
// TODO: Add benchmark tests.
type ArrayWriter struct {
	pBuf, outBuf *bytes.Buffer

	lBuf []byte

	Into io.WriteCloser
}

// NewArrayWriter constructs a new ArrayWriter over the given writer.
func NewArrayWriter(over io.WriteCloser) ArrayWriter {
	var (
		// Page buffer, output buffer
		pBuf, outBuf = new(bytes.Buffer), new(bytes.Buffer)

		// Line buffer plus 1 for terminal newline.
		lbb  [MaxWidth + 1]byte
		lBuf = lbb[:]
	)

	return ArrayWriter{
		pBuf:   pBuf,
		outBuf: outBuf,
		lBuf:   lBuf,
		Into:   over,
	}
}

// Flush flushes any remaining buffered contents to a.Into.
func (a ArrayWriter) Flush() error {
	defer a.outBuf.Reset()
	_, err := io.Copy(a.Into, a.outBuf)
	return errors.Wrap(err, "flushing buffer")
}

// Close calls Flush and then closes the underlying WriteCloser if
// successful.  It may not be used after this.
func (a ArrayWriter) Close() error {
	if err := a.Flush(); err != nil {
		return err
	}
	return a.Into.Close()
}

// Write implements io.Writer on ArrayWriter for the purpose of Copy
// from an io.WriterTo.
func (a ArrayWriter) Write(some []byte) (int, error) {
	// TODO: Make this more efficient (don't allocate a buffer.)
	// TODO: Don't always flush, etc.
	buf := bytes.NewBuffer(some)
	n, err := a.ReadFrom(buf)
	return int(n), err
}

// ReadFrom consumes bytes from the given Reader, translating them into
// hex literals formatted for a C++ byte array literal.
//
// The returned count is the number of bytes consumed from the Reader.
func (a ArrayWriter) ReadFrom(some io.Reader) (total int64, err error) {
	var (
		pb = a.pBuf
		lb = a.lBuf

		done bool
	)

	for !done {
		pb.Reset()
		n, err := io.CopyN(pb, some, PageSize)
		total += n
		switch err {
		case nil:
		case io.EOF:
			// Reached end of input.
			done = true

		default:
			return total, errors.Wrap(err, "buffering")
		}

		from := pb.Bytes()
		for iter := 0; iter < len(from); iter += 11 {
			// Loop over the page, writing out lines of hex.
			// One line of 72 ASCII characters can represent
			// up to 11 bytes of the input page.
			//
			// This could be made more space-efficient by
			// encoding as a base64 string and then decoding
			// in the C++ class, but that seems unnecessary.
			begin, end := iter, lesser(len(from), iter+11)
			lb = lb[0 : 6*(end-begin+1)+1]

			var j int
			for ; j < 11 && iter+j < end; j++ {
				// Loop over each byte in the line of
				// output.

				// Each will consume "0xYY,", plus a
				// comment to represent it in ASCII.
				cb := j * 5

				lb[cb], lb[cb+1] = '0', 'x'

				// This always copies 2 ASCII symbols.
				_ = hex.Encode(
					lb[cb+2:cb+4],
					from[begin+j:begin+j+1],
				)
				lb[cb+4] = ','
			}

			// After encoding j bytes, copy out the comment.
			// j*6+6 = j*5 + len(" // |") + j + len("|\n")
			renderASCIIComment(lb[j*5:j*6+7], from[begin:end])

			// Write to *bytes.Buffer never fails.
			_, _ = a.outBuf.Write(lb[:j*6+7])

			// Flush outBuf every <page size>.
			if a.outBuf.Len() >= PageSize {
				if err = a.Flush(); err != nil {
					return total, err
				}
			}
		}
	}

	return
}

func renderASCIIComment(into, from []byte) {
	copy(into, []byte{' ', '/', '/', ' ', '|'})

	for i, c := range from {
		// TODO: Try to parse as runes?  Check for UTF-8?
		switch {
		case !strconv.IsPrint(rune(c)), unicode.IsSpace(rune(c)):
			into[5+i] = '.'
		default:
			into[5+i] = c
		}
	}

	into[len(from)+5], into[len(from)+6] = '|', '\n'
}

func lesser(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
