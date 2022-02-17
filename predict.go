package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	gsc "github.com/fumiama/go-setu-class"
	"github.com/fumiama/loliana/lolicon"
	"github.com/sirupsen/logrus"
)

const (
	eropath   = "./ero.pt"
	norpath   = "./nor.pt"
	loliurl18 = "https://api.lolicon.app/setu/v2?r18=2&proxy=null"
	loliurlnm = "https://api.lolicon.app/setu/v2?proxy=null"
)

var (
	eroindex int
	norindex int
	// P站特殊客户端
	client  = &http.Client{}
	items   []*lolicon.Item
	itemsmu sync.Mutex
)

func init() {
	eroindex = gsc.LoadModule(eropath)
	norindex = gsc.LoadModule(norpath)
}

// getloliurl return url is18
func getloliurl(hasr18 bool) (string, bool) {
	var link string
	if hasr18 {
		link = loliurl18
	} else {
		link = loliurlnm
	}
	// 网络请求
	resp, err := http.Get(link)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	var item lolicon.Item
	err = json.NewDecoder(resp.Body).Decode(&item)
	if err != nil {
		return "", false
	}
	itemsmu.Lock()
	items = append(items, &item)
	itemsmu.Unlock()
	return item.Original, item.R18
}

// predicturl return class dhash
func predicturl(url string, loli bool, newcls bool, hasr18 bool, nopredict bool) (int, string, []byte) {
	var r18 bool
	if loli {
		url, r18 = getloliurl(hasr18)
	} else {
		r18 = false
	}
	var resp *http.Response
	var err error
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
