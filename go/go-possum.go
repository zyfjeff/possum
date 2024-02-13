package possum

import (
	"github.com/anacrolix/generics"
	possumC "github.com/anacrolix/possum/go/cpossum"
	"io"
	"io/fs"
	"time"
)

type Handle struct {
	cHandle *possumC.Handle
}

func Open(dir string) (*Handle, error) {
	cHandle := possumC.NewHandle(dir)
	return &Handle{cHandle}, nil
}

func (me Handle) Close() error {
	possumC.DropHandle(me.cHandle)
	return nil
}

func (me Handle) SingleStat(key string) (fi FileInfo, ok bool) {
	stat := possumC.SingleStat(me.cHandle, key)
	if !stat.Ok {
		return
	}
	return FileInfo{stat.Value, key}, true
}

func (me Handle) PutBuf(key string, buf []byte) error {
	written, err := possumC.WriteSingleBuf(me.cHandle, key, buf)
	if err != nil {
		return err
	}
	if written != uint(len(buf)) {
		panic("expected an error")
	}
	return err
}

func (me Handle) ListKeys(prefix string) (keys []string, err error) {
	items, err := possumC.HandleListItems(me.cHandle, prefix)
	for _, item := range items {
		keys = append(keys, item.Key)
	}
	return
}

func (me Handle) SingleDelete(key string) (fi generics.Option[FileInfo], err error) {
	stat, err := possumC.SingleDelete(me.cHandle, key)
	if err != nil {
		return
	}
	if !stat.Ok {
		return
	}
	fi.Value = FileInfo{stat.Value, key}
	fi.Ok = true
	return
}

func (me Handle) SingleReadAt(key string, off int64, p []byte) (n int, err error) {
	n, err = possumC.SingleReadAt(me.cHandle, key, p, uint64(off))
	// See the very strict definition of io.ReaderAt.ReadAt.
	if n == 0 && len(p) != 0 && err == nil {
		err = io.EOF
	}
	return
}

func (me Handle) NewReader() (r Reader, err error) {
	r.pc, err = possumC.NewReader(me.cHandle)
	return
}

type Limits = possumC.Limits

func (me Handle) SetInstanceLimits(limits Limits) error {
	return possumC.SetInstanceLimits(me.cHandle, limits)
}

type Reader struct {
	pc possumC.Reader
}

func (r Reader) Add(key string) (v Value, err error) {
	v.c, err = possumC.ReaderAdd(r.pc, key)
	return
}

func (r Reader) Begin() error {
	return possumC.ReaderBegin(r.pc)
}

func (r Reader) End() {
	possumC.ReaderEnd(r.pc)
}

func (r Reader) Close() error {
	// This probably isn't safe to call multiple times.
	r.End()
	return nil
}

func (r Reader) ListItems(prefix string) ([]Item, error) {
	return possumC.ReaderListItems(r.pc, prefix)
}

type Value struct {
	c   possumC.Value
	key string
}

func (v Value) ReadAt(p []byte, off int64) (n int, err error) {
	return possumC.ValueReadAt(v.c, p, off)
}

func (v Value) Stat() FileInfo {
	return FileInfo{possumC.ValueStat(v.c), v.key}
}

type FileInfo struct {
	cStat possumC.Stat
	name  string
}

func (f FileInfo) Name() string {
	return f.name
}

func (f FileInfo) Size() int64 {
	return f.cStat.Size()
}

func (f FileInfo) Mode() fs.FileMode {
	return 0o444
}

func (f FileInfo) ModTime() time.Time {
	return f.cStat.LastUsed()
}

func (f FileInfo) IsDir() bool {
	return false
}

func (f FileInfo) Sys() any {
	return f.cStat
}

type Item = possumC.Item
