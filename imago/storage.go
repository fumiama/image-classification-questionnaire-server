package imago

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

var (
	images = make(map[string][]string)
	mutex  sync.Mutex
)

func Imgexsits(name string) bool {
	index := name[:3]
	tail := name[3:]
	tails, ok := images[index]
	if ok {
		found := false
		for _, t := range tails {
			if tail == t {
				found = true
				break
			}
		}
		return found
	}
	return false
}

func Addimage(name string) {
	index := name[:3]
	tail := name[3:]
	mutex.Lock()
	defer mutex.Unlock()
	if images[index] == nil {
		images[index] = make([]string, 0)
		fmt.Println("[addimage] create index", index, ".")
	}
	images[index] = append(images[index], tail)
	fmt.Println("[addimage] index", index, "append file", tail, ".")
	images["sum"] = append(images["sum"], name)
}

func Saveimgbytes(b []byte, imgdir string, uid string, force bool) string {
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
	dh, err := GetDHashStr(img)
	if err != nil {
		return "\"stat\": \"dherr\""
	}
	if force && Imgexsits(dh) {
		return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(dh) + "\""
	} else {
		for _, name := range images["sum"] {
			diff, err := HammDistance(dh, name)
			if err == nil && diff < 10 { // 认为是一张图片
				fmt.Printf("[saveimg] old %s.\n", name)
				return "\"stat\":\"exist\", \"img\": \"" + url.QueryEscape(name) + "\""
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

func Saveimg(r io.Reader, imgdir string, uid string) string {
	imgbuff := make([]byte, 1024*1024) // 1m
	r.Read(imgbuff)
	return Saveimgbytes(imgbuff, imgdir, uid, false)
}

func Scanimgs(imgdir string) error {
	entry, err := os.ReadDir(imgdir)
	if err != nil {
		return err
	}
	for _, i := range entry {
		if !i.IsDir() {
			name := i.Name()
			if strings.HasSuffix(name, ".webp") {
				Addimage(name[:len(name)-5])
			}
		}
	}
	return nil
}

func namein(name string, list []string) bool {
	in := false
	for _, item := range list {
		if name == item {
			in = true
			break
		}
	}
	return in
}

func Pick(exclude []string) string {
	sum := images["sum"]
	le := len(exclude)
	ls := len(sum)
	if le >= ls {
		return ""
	} else if le == 0 {
		return sum[rand.Intn(len(sum))]
	} else if ls/le > 10 {
		name := sum[rand.Intn(len(sum))]
		for namein(name, exclude) {
			name = sum[rand.Intn(len(sum))]
		}
		return name
	} else {
		for _, n := range sum {
			if !namein(n, exclude) {
				return n
			}
		}
		return ""
	}
}
