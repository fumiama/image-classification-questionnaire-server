package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"time"
	"unsafe"

	base14 "github.com/fumiama/go-base16384"
	"github.com/fumiama/image-classification-questionnaire-server/configo"
	log "github.com/sirupsen/logrus"
)

func getuuid() string {
	stamp := time.Now().Unix()
	timestruct := [3]uintptr{uintptr(unsafe.Pointer(&stamp)), uintptr(8), uintptr(8)}
	timebytes := *(*[]byte)(unsafe.Pointer(&timestruct))
	ima := md5.Sum(timebytes)
	uuid, _ := base14.UTF16be2utf8(base14.Encode(ima[:]))
	return string(uuid)[:6]
}

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

func userexists(uid string) bool {
	_, ok := conf.Users[uid]
	return ok
}

func flushconf() {
	timer := time.NewTicker(time.Minute)
	defer timer.Stop()
	for range timer.C {
		if confchanged {
			err := saveconf()
			if err != nil {
				log.Errorln("[saveconf] error:", err)
			} else {
				log.Println("[saveconf] success.")
			}
			confchanged = false
		} else {
			log.Println("[saveconf] vote not change.")
		}
	}
}

func loadconf(pbfile string) error {
	if exists(pbfile) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					return conf.Unmarshal(data)
				}
			}
			return err1
		}
		return err
	}
	conf.Upload = make(map[string]string)
	conf.Users = make(map[string]*configo.DataVote)
	return nil
}

func saveconf() error {
	data, err := conf.Marshal()
	if err == nil {
		f, err1 := os.OpenFile(configfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 == nil {
			defer f.Close()
			_, err2 := f.Write(data)
			return err2
		}
		return err1
	}
	return err
}

func getkeys(m map[string]uint32) []string {
	// 数组默认长度为map长度,后面append时,不需要重新申请内存和拷贝,效率很高
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getfirst(key string, q url.Values) string {
	keys, ok := q[key]
	if ok {
		return keys[0]
	}
	return ""
}
