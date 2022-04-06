package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"os"
	"time"

	base14 "github.com/fumiama/go-base16384"
	"github.com/fumiama/imago"
	log "github.com/sirupsen/logrus"

	"github.com/fumiama/image-classification-questionnaire-server/configo"
)

func getuuid() string {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(time.Now().Unix()))
	ima := md5.Sum(buf[:])
	uuid, _ := base14.UTF16BE2UTF8(base14.Encode(ima[:]))
	return imago.BytesToString(uuid)[:6]
}

// u82int 字节数(大端)组转成int(无符号的)
func u82int(s string) (int, error) {
	b, err1 := base14.UTF82UTF16BE(imago.StringToBytes(s))
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
			log.Debugln("[saveconf] vote not change.")
		}
	}
}

func loadconf() error {
	data, err := storage.GetConf()
	if err != nil || len(data) == 0 {
		conf.Upload = make(map[string]string)
		conf.Users = make(map[string]*configo.DataVote)
		return nil
	}
	return conf.Unmarshal(data)
}

func saveconf() error {
	data, err := conf.Marshal()
	if err == nil {
		err = storage.SaveConf(data)
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

func getfirst(key string, q *url.Values) string {
	keys, ok := (*q)[key]
	if ok {
		log.Debugln("[getfirst] get query", key, "=", keys[0], ".")
		return keys[0]
	}
	log.Debugln("[getfirst]", key, "has no query.")
	return ""
}

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func methodis(m string, resp http.ResponseWriter, req *http.Request) bool {
	log.Debugf("[methodis] %v from %v.", req.Method, getIP(req))
	if req.Method != m {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
