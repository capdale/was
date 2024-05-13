package test

import (
	"os"
	"path"
)

type TmpDir struct {
	path string
}

func NewTmpDir(pattern string) TmpDir {
	dir, _ := os.MkdirTemp("", pattern)
	return TmpDir{
		path: dir,
	}
}

func (c TmpDir) Close() error {
	return c.Remove()
}

func (c TmpDir) Remove() error {
	return os.RemoveAll(c.path)
}

func (c TmpDir) Join(elem ...string) string {
	return path.Join(append([]string{c.path}, elem...)...)
}
