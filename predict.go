package main

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	gsc "github.com/fumiama/go-setu-class"
	"github.com/fumiama/imago"
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
	// P站 无污染 IP 地址
	IPTables = map[string]string{
		"pixiv.net":   "210.140.131.223:443",
		"i.pximg.net": "210.140.92.142:443",
	}
	// P站特殊客户端
	client = &http.Client{
		// 解决中国大陆无法访问的问题
		Transport: &http.Transport{
			DisableKeepAlives: false,
			// 隐藏 sni 标志
			TLSClientConfig: &tls.Config{
				ServerName:         "-",
				InsecureSkipVerify: true,
			},
			// 更改 dns
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("tcp", IPTables["i.pximg.net"])
			},
		},
	}
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}
	s := imago.Bytes2str(body)
	logrus.Println("[getloli] get body.")
	r := strings.Index(s, "\"r18\":")
	is18 := s[r+6:r+10] == "true"
	logrus.Println("[getloli] is18:", is18, ".")
	s = s[strings.Index(s, "\"urls\":{\"original\":\"")+20:]
	s = s[:strings.Index(s, "\"}")]
	logrus.Println("[getloli] url:", s, ".")
	return s, is18
}

// predicturl return class data dhash fullpath
func predicturl(url string, loli bool, newcls bool, hasr18 bool, nopredict bool) (int, string, string) {
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
		return -1, "", ""
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("[predicturl] read body error:", err, ".")
		return -2, "", ""
	}
	var imagetarget string
	if loli {
		imagetarget = imgdir
	} else {
		imagetarget = custimgdir
	}
	_, dh := imago.Saveimgbytes(data, imagetarget, true, 0)
	if dh == "" {
		logrus.Errorln("[predicturl] get dhash error:", err, ".")
		return -3, dh, ""
	}
	var p int
	filefullpath := imagetarget + dh + ".webp"
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
	return p, dh, filefullpath
}
