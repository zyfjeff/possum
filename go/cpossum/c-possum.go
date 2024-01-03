package possumC

// #cgo LDFLAGS: -lpossum
// #include "../../possum.h"
import "C"
import (
	"errors"
	"github.com/anacrolix/generics"
	"io/fs"
	"math"
	"runtime"
	"time"
	"unsafe"
)

func mapError(err uint32) error {
	switch err {
	case C.NoError:
		return nil
	case C.NoSuchKey:
		return fs.ErrNotExist
	default:
		panic(err)
	}
}

type Stat = C.PossumStat

func (me Stat) LastUsed() time.Time {
	ts := me.last_used
	return time.Unix(int64(ts.secs), int64(ts.nanos))
}

func (me Stat) Size() int64 {
	return int64(me.size)
}

type Handle = C.Handle

func NewHandle(dir string) *Handle {
	cDir := C.CString(dir)
	defer C.free(unsafe.Pointer(cDir))
	handle := C.possum_new(cDir)
	return handle
}

func DropHandle(handle *Handle) {
	C.possum_drop(handle)
}

func SingleStat(handle *Handle, key string) (opt generics.Option[Stat]) {
	opt.Ok = bool(C.possum_single_stat(handle, (*C.char)(unsafe.Pointer(unsafe.StringData(key))), C.size_t(len(key)), &opt.Value))
	return
}

func WriteSingleBuf(handle *Handle, key string, buf []byte) (written uint, err error) {
	written = uint(C.possum_single_write_buf(
		handle,
		(*C.char)(unsafe.Pointer(unsafe.StringData(key))),
		C.size_t(len(key)),
		(*C.uchar)(unsafe.SliceData(buf)),
		C.size_t(len(buf)),
	))
	if written == math.MaxUint {
		err = errors.New("unknown possum error")
	}
	return
}

func goListItems(items *C.PossumItem, itemsLen C.size_t) (goItems []Item) {
	itemsSlice := unsafe.Slice(items, uint(itemsLen))
	goItems = make([]Item, itemsLen)
	for i, from := range itemsSlice {
		to := &goItems[i]
		to.Key = C.GoStringN(
			(*C.char)(from.key.ptr),
			C.int(from.key.size),
		)
		C.free(unsafe.Pointer(from.key.ptr))
		to.Stat = from.stat
	}
	C.free(unsafe.Pointer(items))
	return
}

func HandleListItems(handle *Handle, prefix string) (items []Item, err error) {
	var cItems *C.PossumItem
	var itemsLen C.size_t
	err = mapError(C.possum_list_items(
		handle,
		BufFromString(prefix),
		&cItems, &itemsLen))
	if err != nil {
		return
	}
	items = goListItems(cItems, itemsLen)
	return
}

func SingleReadAt(handle *Handle, key string, buf []byte, offset uint64) (n int, err error) {
	var nByte C.size_t = C.size_t(len(buf))
	err = mapError(C.possum_single_readat(
		handle,
		(*C.char)(unsafe.Pointer(unsafe.StringData(key))),
		C.size_t(len(key)),
		(*C.uchar)(unsafe.SliceData(buf)),
		&nByte,
		C.uint64_t(offset),
	))
	n = int(nByte)
	return
}

type Reader = *C.PossumReader

func NewReader(handle *Handle) (r Reader, err error) {
	err = mapError(C.possum_reader_new(handle, &r))
	return
}

func BufFromString(s string) C.PossumBuf {
	return C.PossumBuf{(*C.char)(unsafe.Pointer(unsafe.StringData(s))), C.size_t(len(s))}
}

func BufFromBytes(b []byte) C.PossumBuf {
	return C.PossumBuf{(*C.char)(unsafe.Pointer(unsafe.SliceData(b))), C.size_t(len(b))}
}

func ReaderAdd(r Reader, key string) (v Value, err error) {
	err = mapError(C.possum_reader_add(r, BufFromString(key), &v))
	return
}

func ReaderBegin(r Reader) error {
	return mapError(C.possum_reader_begin(r))
}

func ReaderEnd(r Reader) error {
	return mapError(C.possum_reader_end(r))
}

func ReaderListItems(r Reader, prefix string) (items []Item, err error) {
	var cItems *C.PossumItem
	var itemsLen C.size_t
	err = mapError(C.possum_reader_list_items(r, BufFromString(prefix), &cItems, &itemsLen))
	if err != nil {
		return
	}
	items = goListItems(cItems, itemsLen)
	return
}

type Value = *C.PossumValue

func ValueReadAt(v Value, buf []byte, offset int64) (n int, err error) {
	pBuf := BufFromBytes(buf)
	var pin runtime.Pinner
	pin.Pin(&pBuf)
	pin.Pin(pBuf.ptr)
	defer pin.Unpin()
	err = mapError(C.possum_value_read_at(v, &pBuf, C.uint64_t(offset)))
	n = int(pBuf.size)
	return
}

type Item struct {
	Key string
	Stat
}