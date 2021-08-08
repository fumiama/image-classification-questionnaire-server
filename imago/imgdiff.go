// Package imago 图片处理相关
package imago

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"unsafe"

	"github.com/corona10/goimagehash"
	base14 "github.com/fumiama/go-base16384"
)

var lastchar = "㴁"

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
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

func decodeDHash(imgname string) *goimagehash.ImageHash {
	b, err := base14.UTF82utf16be(str2bytes(imgname + lastchar))
	if err == nil {
		dhb := base14.Decode(b)
		if dhb != nil {
			dh, err1 := bytes82uint64(dhb)
			base14.Free(dhb)
			if err1 == nil {
				return goimagehash.NewImageHash(dh, goimagehash.DHash)
			}
		}

	}
	return nil
}

func HammDistance(img1 string, img2 string) (int, error) {
	b1 := decodeDHash(img1)
	b2 := decodeDHash(img2)
	return b1.Distance(b2)
}

func GetDHashStr(img image.Image) (string, error) {
	dh, err := goimagehash.DifferenceHash(img)
	if err == nil {
		data := make([]byte, 8)
		binary.BigEndian.PutUint64(data, dh.GetHash())
		e := base14.Encode(data)
		b, _ := base14.UTF16be2utf8(e)
		base14.Free(e)
		return string(b)[:15], nil
	}
	return "", err
}
