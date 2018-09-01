package gen

import (
	"fmt"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/synapse-garden/phx/fs"
	"github.com/synapse-garden/phx/gen/compress"
	"github.com/synapse-garden/phx/gen/cpp"

	"github.com/pkg/errors"
)

// Gen uses Operate to process files in the FS given as From, and copies
// its output to To after processing is completed successfully.  It uses
// a temporary buffer for staging before completion.
type Gen struct {
	From, To fs.FS
	compress.Level

	// TODO: Verbosity
}

// Operate processes files as in the description of the type.
func (g Gen) Operate() error {
	// TODO: Describe pipelines with a graph file.
	// TODO: Generate and check resource manifest for changes.
	fis, err := g.From.ReadDir("")
	if err != nil {
		return errors.Wrapf(err, "reading %s", g.From)
	}

	// In workers, open each file, zip and translate it into a
	// static array, and close it.  When each is done, it should be
	// in the tmp destination.  After they're all done, move them
	// all into the target destination.

	var (
		jobs, dones, kill, errs = MakeChans()

		tmpFS   = fs.MakeSyncMem()
		maker   = compress.LZ4Maker{g.Level}
		encoder = cpp.PrepareTarget(tmpFS, maker)
	)

	for i := 0; i < runtime.NumCPU(); i++ {
		// TODO: Use real tmpdir for very large resources.
		go Work{
			from:    g.From,
			Jobs:    jobs,
			Done:    dones,
			Kill:    kill,
			Errs:    errs,
			Encoder: encoder,
		}.Run()
	}

	go func() {
		for _, fi := range fis {
			// TODO: Check for nested resource dirs.
			// if fi.IsDir() { ... }
			jobs <- Job{fi.Name()}
		}

		close(jobs)
	}()

	tw := new(tabwriter.Writer)
	tw.Init(os.Stdout, 0, 8, 0, '\t', 0)

	for i := 0; i < len(fis); i++ {
		select {
		case err := <-errs:
			close(kill)
			return err

		case d := <-dones:
			fmt.Fprintf(tw, "%s:\t%s\n", d.Name, renderSize(d.Size))
		}
	}

	tw.Flush()

	if err := encoder.Finalize(); err != nil {
		// Do any last synchronous cleanup the Encoder requires.
		return errors.Wrap(err, "finalizing Encoder")
	}

	// All finished tmpfiles are now in the tmp destination and
	// shall be moved over to the target.

	tmpFis, err := tmpFS.ReadDir("")
	if err != nil {
		return errors.Wrap(err, "reading tempdir")
	}

	// TODO: Make this concurrent.
	for _, fi := range tmpFis {
		name := fi.Name()
		if err := fs.Move(tmpFS, g.To, name, name); err != nil {
			return errors.Wrapf(err, "finalizing %s", name)
		}
	}

	// TODO: If we used a real tmpdir, remove it now.

	return nil
}

// Size constants.
const (
	KB = 2 << 9
	MB = 2 << 19
	GB = 2 << 29
)

func renderSize(byteLen int64) string {
	switch {
	case byteLen < KB:
		return fmt.Sprintf("%d B", byteLen)
	case byteLen < MB:
		return fmt.Sprintf("%.2f KB", float64(byteLen)/KB)
	case byteLen < GB:
		return fmt.Sprintf("%.2f MB", float64(byteLen)/MB)
	default:
		return fmt.Sprintf("%.2f GB", float64(byteLen)/GB)
	}
}
