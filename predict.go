package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"

	base14 "github.com/fumiama/go-base16384"
	gsc "github.com/fumiama/go-setu-class"
	"github.com/fumiama/imago"
	"github.com/fumiama/loliana/database"
	"github.com/fumiama/loliana/lolicon"
	"github.com/sirupsen/logrus"
)

const (
	eropath   = "./ero.pt"
	norpath   = "./nor.pt"
	loliurl18 = "https://api.lolicon.app/setu/v2?r18=2&proxy=null"
	loliurlnm = "https://api.lolicon.app/setu/v2?proxy=null"
)

type loliresult struct {
	Error string `json:"error"`
	Data  []struct {
		Pid        int      `json:"pid"`
		P          int      `json:"p"`
		UID        int      `json:"uid"`
		Title      string   `json:"title"`
		Author     string   `json:"author"`
		R18        bool     `json:"r18"`
		Width      int      `json:"width"`
		Height     int      `json:"height"`
		Tags       []string `json:"tags"`
		Ext        string   `json:"ext"`
		UploadDate int64    `json:"uploadDate"`
		Urls       struct {
			Original string `json:"original"`
		} `json:"urls"`
	} `json:"data"`
}

var (
	eroindex int
	norindex int
	// P站特殊客户端
	client      = &http.Client{}
	items       []database.Picture
	tags        = make(map[string]database.Tag)
	itemsmu     sync.Mutex
	pidpreg     = regexp.MustCompile(`\d+_p\d+`)
	datepathreg = regexp.MustCompile(`\d{4}/\d{2}/\d{2}/\d{2}/\d{2}/\d{2}`)
)

func init() {
	eroindex = gsc.LoadModule(eropath)
	norindex = gsc.LoadModule(norpath)
}

func getloliurl(hasr18 bool) (*lolicon.Item, error) {
	var link string
	if hasr18 {
		link = loliurl18
	} else {
		link = loliurlnm
	}
	// 网络请求
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r loliresult
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}
	i := r.Data[0]
	return &lolicon.Item{
		Pid:      i.Pid,
		P:        i.P,
		UID:      i.UID,
		Width:    i.Width,
		Height:   i.Height,
		Title:    i.Title,
		Author:   i.Author,
		R18:      i.R18,
		Tags:     i.Tags,
		Ext:      i.Ext,
		Original: i.Urls.Original,
	}, nil
}

// predicturl return class dhash
func predicturl(url string, loli bool, newcls bool, hasr18 bool, nopredict bool) (int, string, []byte) {
	var r18 bool
	var item *lolicon.Item
	var resp *http.Response
	var err error
	if loli {
		item, err = getloliurl(hasr18)
		if err != nil {
			return -8, "", nil
		}
		url = item.Original
	}
	if loli {
		// 网络请求
		request, _ := http.NewRequest("GET", url, nil)
		request.Header.Set("Host", "i.pximg.net")
		request.Header.Set("Referer", "https://www.pixiv.net/")
		request.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
		resp, err = client.Do(request)
	} else {
		resp, err = http.Get(url)
	}
	if err != nil {
		logrus.Errorln("[predicturl] get url error:", err, ".")
		return -1, "", nil
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("[predicturl] read body error:", err, ".")
		return -2, "", nil
	}
	var imagetarget string
	if loli {
		imagetarget = "img"
	} else {
		imagetarget = "cust"
	}
	stat, dh := storage.SaveImgBytes(data, imagetarget, true, 0)
	logrus.Println("[predicturl]", stat)
	if dh == "" {
		logrus.Errorln("[predicturl] get dhash error:", err, ".")
		return -3, dh, nil
	}
	logrus.Infoln("[predicturl] get dhash :", dh, ".")
	m := md5.Sum(data)
	ms := hex.EncodeToString(m[:])
	logrus.Infoln("[predicturl] get md6 :", ms, ".")
	var p int
	filefullpath := cachedir + "/" + dh + ".webp"
	if !exists(cachedir) {
		err = os.MkdirAll(cachedir, 0755)
		if err != nil {
			logrus.Errorln("[predicturl]", err)
			return -4, dh, nil
		}
	}
	if !exists(filefullpath) {
		data, err = storage.GetImgBytes(imagetarget, dh+".webp")
		if err != nil {
			if !loli {
				data, err = storage.GetImgBytes("img", dh+".webp")
				if err == nil {
					goto SAVECACHE
				}
			}
			logrus.Errorln("[predicturl]", err)
			return -5, dh, nil
		}
	SAVECACHE:
		err = os.WriteFile(filefullpath, data, 0644)
		if err != nil {
			logrus.Errorln("[predicturl]", err)
			return -6, dh, data
		}
		ts, _ := json.Marshal(item.Tags)
		pidp := pidpreg.FindString(item.Original)
		itemsmu.Lock()
		items = append(items, database.Picture{
			PidP:     pidp,
			UID:      item.UID,
			Width:    item.Width,
			Height:   item.Height,
			Title:    item.Title,
			Author:   item.Author,
			R18:      item.R18,
			Tags:     imago.BytesToString(ts),
			Ext:      item.Ext,
			DatePath: datepathreg.FindString(item.Original),
		})
		for _, tag := range item.Tags {
			name, err := base14.UTF16be2utf8(base14.EncodeString(tag))
			if err != nil {
				logrus.Errorln("encode tag", tag, "error:", err)
				continue
			}
			t := imago.BytesToString(name)
			tags[t] = database.Tag{
				PidP: pidp,
				UID:  item.UID,
			}
		}
		itemsmu.Unlock()
	} else {
		data, err = os.ReadFile(filefullpath)
		if err != nil {
			logrus.Errorln("[predicturl]", err)
			return -7, dh, nil
		}
	}
	logrus.Infoln("[predicturl] file path:", filefullpath, ".")
	if !nopredict {
		p = gsc.PredictFile(filefullpath, eroindex)
		logrus.Infoln("[predicturl] ero:", p, ".")
		if newcls {
			n := gsc.PredictFile(filefullpath, norindex)
			logrus.Infoln("[predicturl] nor:", n, ".")
			if p > 4 && n > 3 && n < 6 {
				p += n - 3
			} else if n > 4 && p > 2 {
				p = n + p - 4
			} else if n > 3 && p > 1 {
				p = n
			} else if n == 0 && p > 0 && p < 3 {
				p -= 1
			}
			if p > 8 {
				p = 8
			}
		}
	}
	logrus.Infoln("[predicturl] loli mae:", p, ".")
	if loli {
		if r18 {
			if newcls {
				if p < 6 {
					p = 7
				}
			} else if p < 5 {
				p = 5
			}
		} else {
			if newcls {
				if p > 5 {
					p = 5
				}
			} else if p > 4 {
				p = 4
			}
		}
	}
	logrus.Infoln("[predicturl] loli usiro:", p, ".")
	return p, dh, data
}
