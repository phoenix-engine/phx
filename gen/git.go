package gen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// GitModule is an optional job which can be added to a Gen pipeline to
// clone a given Git repository into the local repo.
type GitModule struct {
	Overwrite bool

	// Remote specifies the full path to the remote, including e.g.
	// protocol.  Local is the local name that the repo will be
	// cloned into.  Revision (optional) defines a specific revision
	// to be checked out, such as a tag, commit id, or branch.  If
	// Branch is set, it will be used in git submodule commands.
	Remote, Local, Branch, Revision string
}

func runOut(w io.Writer, command string, args ...interface{}) (bool, error) {
	cc := fmt.Sprintf(command, args...)
	elements := strings.Fields(cc)
	proc := exec.Command(elements[0], elements[1:]...)
	proc.Stdout = w

	err := errors.Wrapf(proc.Run(), "running %v", cc)
	success := proc.ProcessState.Success()

	switch errors.Cause(err).(type) {
	case nil:
	case *exec.ExitError:
	default:
		return success, err
	}
	return success, nil
}

func run(command string, args ...interface{}) (bool, error) {
	return runOut(os.Stdout, command, args...)
}

// Operate checks for the presence of the Git submodule in the given
// relative path.  If the working directory is not in a Git repo, the
// GitModule is cloned directly into the target root.
//
// If the working directory is in a Git repo, Operate checks for the
// GitModule as a submodule of the Git repo, in the given path.  If it
// is not a submodule in the given path, or if g.Overwrite is true, the
// GitModule is added there as a git submodule using git from the shell.
//
// If it is present, and Overwrite is false, it is ignored.
//
// TODO: Check for a particular revision.
func (g GitModule) Operate(root string) error {
	outpath := filepath.Join(root, g.Local)

	// Are we in a git repo?
	if is, err := runOut(nil, "git rev-parse --is-inside-work-tree"); err != nil {
		return errors.Wrap(err, "creating git status check")
	} else if !is {
		// No.  Clone the repo into the target root.
		ok, err := run("git clone %s %s", g.Remote, outpath)
		if err != nil {
			return errors.Wrapf(err, "cloning git remote %s", g.Remote)
		}
		if !ok {
			return errors.Errorf("cloning git remote %s failed", g.Remote)
		}
		return nil
	}

	// Yes, we are in a git repo.

	// Is the desired GitModule already a submodule?
	var buf bytes.Buffer
	isSubm, err := runOut(&buf, "git submodule status %s", outpath)
	if err != nil {
		return errors.Wrapf(err, "checking git submodule %s", outpath)
	} else if isSubm {
		// Yes, check whether its state matches expected.
		return nil
	}

	// No, add it as a submodule.

	// If it has a branch set, use that branch.
	var b string
	if gb := g.Branch; gb != "" {
		b = "-b " + gb + " "
	}

	ok, err := run("git submodule add %s%s %s", b, g.Remote, outpath)
	if err != nil {
		msg := "adding submodule %s from %s"
		return errors.Wrapf(err, msg, outpath, g.Remote)
	} else if ok {
		// The submodule was added successfully.
		return nil
	}

	return errors.Errorf("failed to add submodule %s from %s", outpath, g.Remote)
}
