package fs

import (
	"bytes"
	"io"
	"os"
	"sort"
	"sync"

	kfs "github.com/kr/fs"
)

type BufCloser struct{ *bytes.Buffer }

func (b BufCloser) Close() error { return nil }

type Mem struct {
	kfs.FileSystem

	bufs map[string]BufCloser
}

type MemInfo struct {
	os.FileInfo
	name string
}

func (m MemInfo) Name() string {
	return m.name
}

type SyncMem struct {
	*sync.RWMutex
	Mem
}

func MakeMem() Mem { return Mem{bufs: make(map[string]BufCloser)} }
func MakeSyncMem() SyncMem {
	return SyncMem{
		new(sync.RWMutex),
		MakeMem(),
	}
}

func (m Mem) ReadDir(some string) (infos []os.FileInfo, err error) {
	var names []string
	for k := range m.bufs {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, name := range names {
		infos = append(infos, MemInfo{name: name})
	}

	return
}

func (m Mem) Join(subs ...string) string {
	return Real{}.Join(subs...)
}

func (m Mem) Split(some string) (parent, rest string) {
	return Real{}.Split(some)
}

func (m Mem) Open(path string) (io.ReadCloser, error) {
	if buf, ok := m.bufs[path]; ok {
		return buf, nil
	}
	return nil, os.ErrNotExist
}

func (m Mem) Create(path string) (io.WriteCloser, error) {
	if _, ok := m.bufs[path]; ok {
		return nil, os.ErrExist
	}

	buf := BufCloser{new(bytes.Buffer)}
	m.bufs[path] = buf
	return buf, nil
}

func (m Mem) Move(from, to string) error {
	bf, ok := m.bufs[from]
	_, ok2 := m.bufs[to]

	switch {
	case !ok:
		return os.ErrNotExist
	case ok2:
		return os.ErrExist
	}

	m.bufs[to] = bf
	delete(m.bufs, from)

	return nil
}

func (s SyncMem) Open(which string) (io.ReadCloser, error) {
	s.RLock()
	defer s.RUnlock()
	return s.Mem.Open(which)
}

func (s SyncMem) Create(which string) (io.WriteCloser, error) {
	s.Lock()
	defer s.Unlock()
	return s.Mem.Create(which)
}

func (s SyncMem) Move(from, to string) error {
	s.Lock()
	defer s.Unlock()
	return s.Mem.Move(from, to)
}
