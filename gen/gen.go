package gen

import (
	"runtime"

	"github.com/synapse-garden/phx/fs"

	"github.com/golang/snappy"
	"github.com/pkg/errors"
)

type Gen struct{ From, To fs.FS }

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
			from: g.From,
			tmp:  tmpFS,
			Jobs: jobs,
			Done: dones,
			Kill: kill,
			Errs: errs,
			Snap: snappy.NewBufferedWriter(nil),
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

		case <-dones:
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
