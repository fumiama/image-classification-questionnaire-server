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
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	base14 "github.com/fumiama/go-base16384"
	"github.com/fumiama/image-classification-questionnaire-server/imago"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

// unpack 从 HTTP 请求 req 的参数中提取数据填充到 ptr 指向结构体的各个字段
func unpack(req *http.Request, ptr interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	// 创建字段映射表，键为有效名称
	fields := make(map[string]reflect.Value)
	v := reflect.ValueOf(ptr).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i)
		tag := fieldInfo.Tag
		name := tag.Get("http")
		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		fields[name] = v.Field(i)
	}

	// 对请求中的每个参数更新结构体中对应的字段
	for name, values := range req.Form {
		f := fields[name]
		if !f.IsValid() {
			continue // 忽略不能识别的 HTTP 参数
		}

		for _, value := range values {
			if f.Kind() == reflect.Slice {
				elem := reflect.New(f.Type().Elem()).Elem()
				if err := populate(elem, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
				f.Set(reflect.Append(f, elem))
			} else {
				if err := populate(f, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
			}
		}
	}
	return nil
}

func populate(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)

	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}
	return nil
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

func saveimg(r io.Reader, uid string) string {
	img, _, err := image.Decode(r)
	iswebp := false
	if err != nil {
		img, err = webp.Decode(r, &decoder.Options{})
		if err == nil {
			iswebp = true
		} else {
			fmt.Println("[saveimg] decode image error.")
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
		c, err := io.Copy(f, r)
		if err != nil {
			return "\"stat\": \"ioerr\""
		}
		fmt.Printf("[saveimg] save %d bytes.\n", c)
	}
	fmt.Printf("[saveimg] new %s.\n", dh)
	return "\"stat\":\"success\", \"img\": \"" + url.QueryEscape(dh) + "\""
}
