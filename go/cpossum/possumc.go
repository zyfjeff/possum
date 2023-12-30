package possumC

// #cgo LDFLAGS: -lpossum
// #include "../../possum.h"
import "C"
import (
	"errors"
	"github.com/anacrolix/generics"
	"math"
	"time"
	"unsafe"
)

type Handle = C.Handle

type Stat = C.PossumStat

func (me Stat) LastUsed() time.Time {
	ts := me.last_used
	return time.Unix(int64(ts.secs), int64(ts.nanos))
}

func (me Stat) Size() int64 {
	return int64(me.size)
}

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

func mapError(err uint32) error {
	switch err {
	case C.NoError:
		return nil
	default:
		panic(err)
	}
}

func ListKeys(handle *Handle, prefix string) (keys []string, err error) {
	var items *C.possum_item
	var itemsLen C.size_t
	err = mapError(C.possum_list_keys(
		handle,
		(*C.uchar)(unsafe.StringData(prefix)),
		C.size_t(len(prefix)),
		&items, &itemsLen))
	if err != nil {
		return
	}
	itemsSlice := unsafe.Slice(items, uint(itemsLen))
	keys = make([]string, itemsLen)
	for i, from := range itemsSlice {
		keys[i] = C.GoStringN(
			(*C.char)(from.key),
			C.int(from.key_size),
		)
		C.free(unsafe.Pointer(from.key))
	}
	C.free(unsafe.Pointer(items))
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
