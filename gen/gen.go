package gen

import (
	"fmt"
	"runtime"

	"github.com/synapse-garden/phx/fs"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
)

// Gen uses Operate to process files in the FS given as From, and copies
// its output to To after processing is completed successfully.  It uses
// a temporary buffer for staging before completion.
type Gen struct {
	From, To fs.FS
	Level
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

	jobs, dones, kill, errs := MakeChans()

	tmpFS := fs.MakeSyncMem()

	for i := 0; i < runtime.NumCPU(); i++ {
		// TODO: Use real tmpdir for very large resources.
		go Work{
			from:       g.From,
			tmp:        tmpFS,
			Jobs:       jobs,
			Done:       dones,
			Kill:       kill,
			Errs:       errs,
			Compressor: LZ4{lz4.NewWriter(nil), g.Level},
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

	for i := 0; i < len(fis); i++ {
		select {
		case err := <-errs:
			close(kill)
			return err

		case d := <-dones:
			fmt.Printf("%s: %d\n", d.Name, d.Size)
		}
	}

	// All finished tmpfiles are now in the tmp destination and
	// shall be moved over to the target.

	// TODO: Make this concurrent?
	for _, fi := range fis {
		name := fi.Name()
		if err := fs.Move(tmpFS, g.To, name, name); err != nil {
			return errors.Wrapf(err, "finalizing %s", name)
		}
	}

	return nil
}
