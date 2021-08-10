package imago

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// bytes82uint64 字节数(大端)组转成int(无符号的)
func bytes82uint64(b []byte) (uint64, error) {
	if len(b) == 9 {
		b = b[:7]
	}
	if len(b) == 8 {
		bytesBuffer := bytes.NewBuffer(b)
		var tmp uint64
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return tmp, err
	}
	return 0, fmt.Errorf("%s", "bytes lenth is invaild!")
}
