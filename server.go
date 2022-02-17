package main

import (
	_ "embed" // embed index
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	para "github.com/fumiama/go-hide-param"
	"github.com/fumiama/imago"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/fumiama/image-classification-questionnaire-server/configo"
)

const cachedir = "cache"

//go:embed index_quart.html
var indexdata string

var (
	pwd         int
	conf        configo.Data
	storage     imago.StorageInstance
	confchanged = false
	defuploader = url.QueryEscape("涩酱")
)

func index(resp http.ResponseWriter, req *http.Request) {
	if methodis("GET", resp, req) {
		io.WriteString(resp, indexdata)
	}
}

func signup(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if methodis("GET", resp, req) {
		q := req.URL.Query()
		key, ok := q["key"]
		if !ok {
			http.Error(resp, "400 BAD REQUEST\nInvalid key.", http.StatusBadRequest)
			log.Errorln("[/signup] invalid key.")
		} else {
			keyint, err := strconv.Atoi(key[0])
			if !ok || err != nil {
				http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
				log.Errorln("[/signup] bad request.")
			} else {
				diff := int(time.Now().Unix()) - (keyint ^ pwd)
				if diff < 10 && diff >= 0 {
					uuid := getuuid()
					for conf.Users[uuid] != nil {
						uuid = getuuid()
					}
					conf.Users[uuid] = new(configo.DataVote)
					conf.Users[uuid].Data = make(map[string]uint32)
					log.Println("[/signup] create new map.")
					log.Printf("[/signup] create user: %s.", uuid)
					fmt.Fprintf(resp, "{\"stat\": \"success\", \"id\": \"%s\"}", url.QueryEscape((uuid)))
					confchanged = true
				} else {
					io.WriteString(resp, "{\"stat\": \"wrong\", \"id\": \"null\"}")
					log.Errorln("[/signup] auth failed.")
				}
			}
		}
	}
}

func img(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if methodis("GET", resp, req) {
		q := req.URL.Query()
		path, ok := q["path"]
		if !ok {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			log.Errorln("[/img] bad request.")
		} else {
			if len([]rune(path[0])) == 5 {
				cachefile := cachedir + "/" + path[0] + ".webp"
				if exists(cachefile) {
					http.ServeFile(resp, req, cachefile)
					log.Printf("[/img] serve cached %s.", path[0])
				} else if storage.IsImgExsits(path[0]) {
					data, err := storage.GetImgBytes("img", path[0]+".webp")
					if err != nil {
						http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
					} else {
						resp.Header().Add("Content-Type", "image/webp")
						resp.Write(data)
						log.Printf("[/img] serve %s.", path[0])
					}
				} else {
					resp.WriteHeader(404)
					io.WriteString(resp, "{\"stat\": \"nosuchimg\"}")
					log.Errorf("[/img] %s not found.", path[0])
				}
			} else {
				resp.WriteHeader(404)
				io.WriteString(resp, "{\"stat\": \"invimg\"}")
				log.Errorf("[/img] invalid image path %s.", path[0])
			}
		}
	}
}

func upload(resp http.ResponseWriter, req *http.Request) {
	// 检查是否POST请求
	if methodis("POST", resp, req) {
		// 检查uid
		q := req.URL.Query()
		uid, ok := q["uuid"]
		if !ok {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			log.Errorln("[/upload] bad request.")
		} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
			if req.ContentLength <= 0 {
				io.WriteString(resp, "{\"stat\": \"emptybody\"}")
				log.Errorln("[/upload] invalid content length.")
			} else {
				buf := make([]byte, req.ContentLength)
				result := make([]byte, 1, 1024)
				result[0] = '['
				cnt := 0
				var err error
				cl := int(req.ContentLength)
				for n := 0; err == nil && cnt < cl; {
					n, err = req.Body.Read(buf[cnt:])
					cnt += n
				}
				log.Printf("[/upload] receive %v/%v bytes.", cnt, req.ContentLength)
				if err == nil || cnt == cl {
					ret, dh := storage.SaveImgBytes(buf, "img", false, 3)
					result = append(result, '{')
					result = append(result, imago.StringToBytes(ret)...)
					result = append(result, '}')
					conf.Upload[dh] = uid[0]
					confchanged = true
					log.Infof("[/upload] user %v save image %v.", uid[0], dh)
				} else {
					log.Errorf("[/upload] receive body error: %v", err)
				}
				result = append(result, ']')
				io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.BytesToString(result)+"}")
			}
		} else {
			io.WriteString(resp, "{\"stat\": \"noid\"}")
			log.Errorln("[/upload] no such user.")
		}
	}
}

func upform(resp http.ResponseWriter, req *http.Request) {
	// 检查是否POST请求
	if methodis("POST", resp, req) {
		// 检查uid
		q := req.URL.Query()
		uid, ok := q["uuid"]
		if !ok {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			log.Errorln("[/upload] bad request.")
		} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
			err := req.ParseMultipartForm(64 * 1024 * 1024)
			if err == nil {
				result := make([]byte, 1, 1024)
				result[0] = '['
				tail := imago.StringToBytes("}, ")
				for _, f := range req.MultipartForm.File["img"] {
					log.Printf("[/upform] receive %v of %v bytes.", f.Filename, f.Size)
					fo, err := f.Open()
					if err == nil {
						ret, dh := storage.SaveImg(fo, "img", 3)
						result = append(result, '{')
						result = append(result, imago.StringToBytes(ret)...)
						result = append(result, tail...)
						conf.Upload[dh] = uid[0]
						confchanged = true
						log.Infof("[/upform] user %v save image %v.", uid[0], dh)
					} else {
						log.Errorf("[/upform] save %v error.", f.Filename)
					}
				}
				result = append(result[:len(result)-2], ']')
				io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.BytesToString(result)+"}")
			} else {
				io.WriteString(resp, "{\"stat\": \"ioerr\"}")
				log.Errorln("[/upform] parse multipart form error.")
			}
		} else {
			io.WriteString(resp, "{\"stat\": \"noid\"}")
			log.Errorln("[/upform] no such user.")
		}
	}
}

func vote(resp http.ResponseWriter, req *http.Request) {
	if methodis("GET", resp, req) {
		q := req.URL.Query()
		uid, ok := q["uuid"]
		img, ok1 := q["img"]
		cls, ok2 := q["class"]
		if !ok || !ok1 || !ok2 {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			log.Errorln("[/vote] bad request.")
		} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
			if len([]rune(img[0])) == 5 && storage.IsImgExsits(img[0]) {
				class, err := strconv.Atoi(cls[0])
				if err == nil && class >= 0 && class <= 7 {
					log.Printf("[/vote] user %s voted %d for %s.", uid[0], class, img[0])
					if conf.Users[uid[0]].Data == nil {
						conf.Users[uid[0]].Data = make(map[string]uint32)
					}
					conf.Users[uid[0]].Data[img[0]] = uint32(class)
					io.WriteString(resp, "{\"stat\": \"success\"}")
					log.Println("[/vote] success.")
					confchanged = true
				} else {
					io.WriteString(resp, "{\"stat\": \"invclass\"}")
					log.Errorln("[/vote] invalid class", class, ".")
				}
			} else {
				io.WriteString(resp, "{\"stat\": \"invimg\"}")
				log.Errorln("[/vote] invalid image", img[0], ".")
			}
		} else {
			io.WriteString(resp, "{\"stat\": \"invid\"}")
			log.Errorln("[/vote] invalid uid", uid[0], ".")
		}
	}
}

func pickof(resp http.ResponseWriter, req *http.Request, isdl bool) {
	if methodis("GET", resp, req) {
		// 检查uid
		q := req.URL.Query()
		uid, ok := q["uuid"]
		if !ok {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			log.Errorln("[pickof] bad request.")
		} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
			exclude := getkeys(conf.Users[uid[0]].Data)
			name := storage.Pick(exclude)
			if name == "" {
				io.WriteString(resp, "{\"stat\": \"nomoreimg\"}")
				log.Errorln("[pickof] no more image.")
			} else if isdl {
				data, err := storage.GetImgBytes("img", name+".webp")
				if err != nil {
					http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
				} else {
					resp.Header().Add("Content-Type", "image/webp")
					resp.Write(data)
					log.Println("[/pickdl]", name, ".")
				}
			} else {
				uploader, ok := conf.Upload[name]
				if !ok {
					uploader = defuploader
				} else {
					uploader = url.QueryEscape(uploader)
				}
				io.WriteString(resp, "{\"stat\": \"success\", \"img\": \""+url.QueryEscape(name)+"\", \"uploader\": \""+uploader+"\"}")
				log.Println("[/pick]", name, ".")
			}
		} else {
			io.WriteString(resp, "{\"stat\": \"noid\"}")
			log.Errorln("[pickof] no such user.")
		}
	}
}

func pick(resp http.ResponseWriter, req *http.Request) {
	pickof(resp, req, false)
}

func pickdl(resp http.ResponseWriter, req *http.Request) {
	pickof(resp, req, true)
}

func dice(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if methodis("GET", resp, req) {
		var loli, noimg, newcls, r18, nopredict bool
		var link string
		if _, err := storage.GetConf(); err != nil {
			http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
			log.Errorln("[/dice] predict error: can't connect to storage.")
			return
		}
		// 检查url
		q := req.URL.Query()
		log.Infoln("[/dice] query:", q, ".")
		link = getfirst("url", &q)
		loli = getfirst("loli", &q) == "true"
		noimg = getfirst("noimg", &q) == "true"
		r18 = getfirst("r18", &q) == "true"
		clsnum := getfirst("class", &q)
		newcls = clsnum == "9"
		nopredict = clsnum == "0"
		c, dh, f := predicturl(link, loli, newcls, r18, nopredict)
		if c >= 0 {
			class := strconv.Itoa(c)
			edh := url.QueryEscape(dh)
			if noimg {
				io.WriteString(resp, "{\"img\": \""+edh+"\", \"class\": \""+class+"\"}")
			} else {
				resp.Header().Add("Class", class)
				resp.Header().Add("DHash", edh)
				resp.Header().Add("Content-Type", "image/webp")
				resp.Write(f)
			}
			runtime.GC()
		} else {
			http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
			log.Errorln("[/dice] predict error:", c, ".")
		}
	}
}

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%time%][%lvl%]: %msg%\n",
	})
	log.SetLevel(log.InfoLevel)
}

func main() {
	arglen := len(os.Args)
	if arglen == 5 || arglen == 6 {
		apiurl := os.Args[2]

		pwd, _ = u82int(os.Args[3])
		para.Hide(3)

		key := os.Args[4]
		storage = imago.NewRemoteStorage(apiurl, key)
		if storage == nil {
			panic("wrong remote para")
		}
		para.Hide(4)

		err := loadconf()
		if err != nil {
			panic(err)
		}
		go flushconf()

		err = storage.ScanImgs("img")
		if err != nil {
			panic(err)
		}

		listener, err := net.Listen("tcp", os.Args[1])
		if err != nil {
			panic(err)
		} else {
			if arglen == 6 {
				uid, err1 := strconv.Atoi(os.Args[5])
				if err1 == nil {
					syscall.Setuid(uid)
					syscall.Setgid(uid)
				} else {
					panic(err1)
				}
			}
			http.HandleFunc("/", index)
			http.HandleFunc("/index.html", index)
			http.HandleFunc("/signup", signup)
			http.HandleFunc("/img", img)
			http.HandleFunc("/upform", upform)
			http.HandleFunc("/upload", upload)
			http.HandleFunc("/vote", vote)
			http.HandleFunc("/pick", pick)
			http.HandleFunc("/pickdl", pickdl)
			http.HandleFunc("/dice", dice)
			// http.Handle("/yuka/", http.StripPrefix("/yuka/", http.FileServer(http.Dir(imgdir))))
			defer func() {
				if err == nil {
					f, err := os.Create(fmt.Sprintf("newloli%d.json", time.Now().Unix()))
					if err == nil {
						json.NewEncoder(f).Encode(&items)
					}
				}
			}()
			log.Error(http.Serve(listener, nil))
		}
	} else {
		fmt.Println("Usage: <listen_addr> <apiurl> <password> <authkey> (userid)")
	}
}
