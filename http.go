package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"strings"

	base14 "github.com/fumiama/go-base16384"
	"github.com/fumiama/image-classification-questionnaire-server/imago"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

var imgbuff = make([]byte, 4*1024*1024) // 4m

// u82int 字节数(大端)组转成int(无符号的)
func u82int(s string) (int, error) {
	b, err1 := base14.UTF82utf16be([]byte(s))
	if err1 != nil {
		return 0, err1
	}
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	switch len(b) {
	case 1:
		var tmp uint8
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	case 2:
		var tmp uint16
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	case 4:
		var tmp uint32
		err := binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp), err
	default:
		return 0, fmt.Errorf("%s", "u82Int bytes lenth is invaild!")
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func saveimgbytes(b []byte, uid string) string {
	r := bytes.NewReader(b)
	img, _, err := image.Decode(r)
	iswebp := false
	if err != nil {
		r.Seek(0, io.SeekStart)
		img, err = webp.Decode(r, &decoder.Options{})
		if err == nil {
			iswebp = true
		} else {
			fmt.Printf("[saveimg] decode image error: %v\n", err)
			return "\"stat\": \"notanimg\""
		}
	}
	dh, err := imago.GetDHashStr(img)
	if err != nil {
		return "\"stat\": \"dherr\""
	}
	entry, err := os.ReadDir(imgdir)
	if err != nil {
		return "\"stat\": \"lserr\""
	}
	for _, i := range entry {
		if !i.IsDir() {
			name := i.Name()
			if strings.HasSuffix(name, ".webp") {
				name = name[:len(name)-5]
				diff, err := imago.HammDistance(dh, name)
				if err == nil && diff < 10 { // 认为是一张图片
					fmt.Printf("[saveimg] old %s.\n", name)
					return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(name) + "\""
				}
			}
		}
	}
	f, err := os.Create(imgdir + dh + ".webp")
	if err != nil {
		return "\"stat\": \"ioerr\""
	}
	defer f.Close()
	if !iswebp {
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
		if err != nil || webp.Encode(f, img, options) != nil {
			return "\"stat\": \"encerr\""
		}
	} else {
		r.Seek(0, io.SeekStart)
		c, err := io.Copy(f, r)
		if err != nil {
			return "\"stat\": \"ioerr\""
		}
		fmt.Printf("[saveimg] save %d bytes.\n", c)
	}
	fmt.Printf("[saveimg] new %s.\n", dh)
	return "\"stat\":\"success\", \"img\": \"" + url.QueryEscape(dh) + "\""
}

func saveimg(r io.Reader, uid string) string {
	r.Read(imgbuff)
	return saveimgbytes(imgbuff, uid)
}
