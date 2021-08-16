package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/fumiama/image-classification-questionnaire-server/configo"
	"github.com/fumiama/imago"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

const indexdata = "<!DOCTYPE html>\n" +
	"<html>\n" +
	"       <head>\n" +
	"               <meta charset=\"utf-8\">\n" +
	"               <title>图像分类问卷调查</title>\n" +
	"               <style>\n" +
	"                       body {\n" +
	"                               font-size: 2vh;\n" +
	"                       }\n" +
	"                       /*\n" +
	"                        *具有缩放效果的按钮样式，父级需要设置font-size\n" +
	"                        *可以通过background-color,border-color覆盖得到不同颜色的按钮\n" +
	"                        **/\n" +
	"                       button {\n" +
	"                               padding: .3em .8em;\n" +
	"                               border: 1px solid #ffb2c5;\n" +
	"                               /*background: #58a linear-gradient(#77a0bb,#58a);*/\n" +
	"                               background: rgba(255, 210, 216, 0.7) linear-gradient(hsla(0,0%,100%,.2),transparent);\n" +
	"                               border-radius: .2em;\n" +
	"                               /*box-shadow: 0 .05em .25em gray;*/\n" +
	"                               box-shadow: 0 .05em .25em rgba(0,0,0,.5);\n" +
	"                               color: black;\n" +
	"                               text-shadow: 0 -.05em .05em #335166;\n" +
	"                               font-size: 125%;\n" +
	"                               line-height: 1.5;\n" +
	"                       }\n" +
	"                       input {\n" +
	"                               padding: .3em .8em;\n" +
	"                               border: 1px solid #ffb2c5;\n" +
	"                               /*background: #58a linear-gradient(#77a0bb,#58a);*/\n" +
	"                               background: rgba(255, 210, 216, 0.7) linear-gradient(hsla(0,0%,100%,.2),transparent);\n" +
	"                               border-radius: .2em;\n" +
	"                               /*box-shadow: 0 .05em .25em gray;*/\n" +
	"                               box-shadow: 0 .05em .25em rgba(0,0,0,.5);\n" +
	"                               color: black;\n" +
	"                               text-shadow: 0 -.05em .05em #335166;\n" +
	"                               font-size: 125%;\n" +
	"                               line-height: 1.5;\n" +
	"                       }\n" +
	"                       div{\n" +
	"                               text-align:center;\n" +
	"                       }\n" +
	"                       .img_box{\n" +
	"                               padding-bottom:100%;\n" +
	"                       }\n" +
	"                       .img_box img{\n" +
	"                               position:fixed;\n" +
	"                               top:0;\n" +
	"                               bottom:0;\n" +
	"                               left:0;\n" +
	"                               right:0;\n" +
	"                               height:100%;\n" +
	"                               margin:auto;\n" +
	"                               z-index: -1;\n" +
	"                               max-height: 100%;\n" +
	"                               max-width: 100%;\n" +
	"                               object-fit: contain;\n" +
	"                               object-position: center;\n" +
	"                               vertical-align: center;\n" +
	"                       }\n" +
	"                       .btn_show_foot{\n" +
	"                               z-index: 1;\n" +
	"                               position: absolute;\n" +
	"                               bottom: 16px;\n" +
	"                               width: 100%;\n" +
	"                       }\n" +
	"               </style>\n" +
	"       </head>\n" +
	"       <body>\n" +
	"               <script>\n" +
	"                       function 增录(名, 值, 日) {\n" +
	"                               var d = new Date();\n" +
	"                               d.setTime(d.getTime()+(日*24*60*60*1000));\n" +
	"                               var 止 = \"expires=\"+d.toGMTString();\n" +
	"                               document.cookie = 名 + \"=\" + 值 + \"; \" + 止;\n" +
	"                       }\n" +
	"                       function 取录(名) {\n" +
	"                               var name = 名 + \"=\";\n" +
	"                               var ca = document.cookie.split(';');\n" +
	"                               for(var i=0; i<ca.length; i++) {\n" +
	"                                       var c = ca[i].trim();\n" +
	"                                       if (c.indexOf(name)==0) return c.substring(name.length,c.length);\n" +
	"                               }\n" +
	"                               return \"\";\n" +
	"                       }\n" +
	"                       我 = 取录(\"uuid\");\n" +
	"                       function 取信(网址, 处理) {\n" +
	"                               var 请求 = new XMLHttpRequest();        //第一步：建立所需的对象\n" +
	"                               请求.open('GET', 网址, true);           //第二步：打开连接,将请求参数写在网址中\n" +
	"                               请求.send();                                            //第三步：发送请求\n" +
	"                               请求.onreadystatechange = function () {\n" +
	"                                       if (请求.readyState == 4 && 请求.status == 200) {\n" +
	"                                               处理(请求.responseText);\n" +
	"                                       }\n" +
	"                               };\n" +
	"                       }\n" +
	"                       function 取随机图(){\n" +
	"                               if(我 == \"\") alert(\"未登录!\");\n" +
	"                               else {\n" +
	"                                       取信(\"/pick?uuid=\" + 我, function rri(t) {\n" +
	"                                               j = JSON.parse(t)\n" +
	"                                               if(j.stat == \"success\") document.getElementById(\"img_display\").src = \"/img?path=\" + j.img;\n" +
	"                                               else if(j.stat == \"nomoreimg\") alert(\"无更多图片!\");\n" +
	"                                               else alert(\"随机失败，请重试\");\n" +
	"                                       });\n" +
	"                               }\n" +
	"                       }\n" +
	"                       function 登录() {\n" +
	"                               入 = prompt(\"请输入用户名，错误的用户名无法加载图片\",\"示例\");\n" +
	"                               if(入 != null) {\n" +
	"                                       if(入.length == 2) {\n" +
	"                                               我 = 入;\n" +
	"                                               增录(\"uuid\", 我, 7);\n" +
	"                                       }\n" +
	"                                       else if(入.length == 0) document.cookie = 我 = \"\";\n" +
	"                               }\n" +
	"                       }\n" +
	"                       function 编码(文) {\n" +
	"                               文 = escape(文.toString()).replace(/\\+/g, \"%2B\");\n" +
	"                               var 配 = 文.match(/(%([0-9A-F]{2}))/gi);\n" +
	"                               if (配) {\n" +
	"                                       for (var 位 = 0; 位 < 配.length; 位++) {\n" +
	"                                               var 码 = 配[位].substring(1,3);\n" +
	"                                               if (parseInt(码, 16) >= 128) {\n" +
	"                                                       文 = 文.replace(配[位], '%u00' + 码);\n" +
	"                                               }\n" +
	"                                       }\n" +
	"                               }\n" +
	"                               文 = 文.replace('%25', '%u0025');\n" +
	"                               return 文;\n" +
	"                       }\n" +
	"                       function 六十(六) {\n" +
	"                       var 长 = 六.length, 串 = new Array(长), 码;\n" +
	"                       for (var 位 = 0; 位 < 长; 位++) {\n" +
	"                               码 = 六.charCodeAt(位);\n" +
	"                               if (48<=码 && 码 < 58) 码 -= 48;\n" +
	"                               else 码 = (码 & 0xdf) - 65 + 10;\n" +
	"                               串[位] = 码;\n" +
	"                       }\n" +
	"                       return 串.reduce(function(和, 余) {\n" +
	"                               和 = 16 * 和 + 余;\n" +
	"                               return 和;\n" +
	"                       }, 0);\n" +
	"                       }\n" +
	"                       function 注册() {\n" +
	"                               if(我 == \"\") {\n" +
	"                                       入 = 编码(prompt(\"请输入密码\"));\n" +
	"                                       入 = 六十(入.substring(2,6) + 入.substring(8, 12));\n" +
	"                                       码 = ((Date.parse(new Date())/1000) ^ 入).toString().padStart(10, \"0\");\n" +
	"                                       取信(\"/signup?key=\" + 码, function rr(t) {\n" +
	"                                               j = JSON.parse(t);\n" +
	"                                               if(j.stat == \"success\") {\n" +
	"                                                       我 = decodeURI(j.id);\n" +
	"                                                       增录(\"uuid\", 我, 7);\n" +
	"                                                       prompt(\"这是您的用户名，请复制好后妥善保存\", 我);\n" +
	"                                               } else alert(\"错误!\");\n" +
	"                                       });\n" +
	"                               }\n" +
	"                       }\n" +
	"                       function 投票(类) {\n" +
	"                               if(我 != \"\") {\n" +
	"                                       图 = document.getElementById(\"img_display\").src;\n" +
	"                                       取信(\"/vote?uuid=\" + 我 + \"&img=\" + 图.substring(图.lastIndexOf('=')+1, 图.length) + \"&class=\" + 类, function rv(t) {\n" +
	"                                               取随机图();\n" +
	"                                       });\n" +
	"                               } else alert(\"请登录!\");\n" +
	"                       }\n" +
	"                       隐 = false;\n" +
	"                       栏 = document.getElementsByTagName(\"div\");\n" +
	"                       function 显隐() {\n" +
	"                               隐 = !隐;\n" +
	"                               栏[0].hidden = 栏[2].hidden = 栏[4].hidden = 隐;\n" +
	"                               document.getElementById(\"btn_hide\").innerText = 隐?\"显示\":\"隐藏\";\n" +
	"                       }\n" +
	"                       function 上传() {\n" +
	"                               document.getElementById(\"upload_form\").action = \"upform?uuid=\" + 我;\n" +
	"                       }\n" +
	"               </script>\n" +
	"               <div>\n" +
	"                       <h1>图像分类问卷调查</h1>\n" +
	"                       <button id = \"btn_rand\" type=\"button\" onclick=\"取随机图()\">随机</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button id = \"btn_lgin\" type=\"button\" onclick=\"登录()\">登录</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button id = \"btn_regi\" type=\"button\" onclick=\"注册()\">注册</button>\n" +
	"               </div>\n" +
	"               <div>\n" +
	"                       <br><br>\n" +
	"                       <button id = \"btn_hide\" type=\"button\" onclick=\"显隐()\">隐藏</button>\n" +
	"               </div>\n" +
	"               <div>\n" +
	"                       <br><br>\n" +
	"                       <form id=\"upload_form\" method=\"post\" formenctype=\"multipart/form-data\" enctype=\"multipart/form-data\">\n" +
	"                               <input type=\"file\" id=\"file\" multiple=\"multiple\" name=\"img\" accept=\".jpg, .jpeg, .png, .webp\">\n" +
	"                               <input type=\"submit\" onclick=\"上传()\">\n" +
	"                       </form>\n" +
	"               </div>\n" +
	"               <div class=\"img_box\">\n" +
	"                       <img id = \"img_display\" src=\"/img?path=嗏蒷篍臑呀\"/>\n" +
	"               </div>\n" +
	"               <div class=\"btn_show_foot\">\n" +
	"                       <button type=\"button\" onclick=\"投票('0')\">0分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('1')\">1分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('2')\">2分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('3')\">4分</button>\n" +
	"                       <br><br>\n" +
	"                       <button type=\"button\" onclick=\"投票('4')\">8分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('5')\">16分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('6')\">32分</button>\n" +
	"                       &nbsp;&nbsp;&nbsp;\n" +
	"                       <button type=\"button\" onclick=\"投票('7')\">64分</button>\n" +
	"               </div>\n" +
	"       </body>\n" +
	"</html>"

var (
	pwd         int
	conf        configo.Data
	imgdir      string
	custimgdir  string
	configfile  string
	confchanged = false
	defuploader = url.QueryEscape("涩酱")
)

func index(resp http.ResponseWriter, req *http.Request) {
	io.WriteString(resp, indexdata)
}

func signup(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if req.Method != "GET" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
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

func img(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if req.Method != "GET" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	q := req.URL.Query()
	path, ok := q["path"]
	if !ok {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		log.Errorln("[/img] bad request.")
	} else {
		if len([]rune(path[0])) == 5 {
			imgpath := imgdir + path[0] + ".webp"
			if imago.Imgexsits(path[0]) {
				http.ServeFile(resp, req, imgpath)
				log.Printf("[/img] serve %s.", path[0])
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

func upload(resp http.ResponseWriter, req *http.Request) {
	// 检查是否POST请求
	if req.Method != "POST" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
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
				ret, dh := imago.Saveimgbytes(buf, imgdir, false, 3)
				result = append(result, '{')
				result = append(result, imago.Str2bytes(ret)...)
				result = append(result, '}')
				conf.Upload[dh] = uid[0]
				confchanged = true
			} else {
				log.Errorf("[/upload] receive body error: %v", err)
			}
			result = append(result, ']')
			log.Infof("[/upload] save image %v.", result)
			io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.Bytes2str(result)+"}")
		}
	} else {
		io.WriteString(resp, "{\"stat\": \"noid\"}")
		log.Errorln("[/upload] no such user.")
	}
}

func upform(resp http.ResponseWriter, req *http.Request) {
	// 检查是否POST请求
	if req.Method != "POST" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
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
			tail := imago.Str2bytes("}, ")
			for _, f := range req.MultipartForm.File["img"] {
				log.Printf("[/upform] receive %v of %v bytes.", f.Filename, f.Size)
				fo, err := f.Open()
				if err == nil {
					ret, dh := imago.Saveimg(fo, imgdir, 3)
					result = append(result, '{')
					result = append(result, imago.Str2bytes(ret)...)
					result = append(result, tail...)
					conf.Upload[dh] = uid[0]
					confchanged = true
				} else {
					log.Errorf("[/upform] save %v error.", f.Filename)
				}
				log.Infof("[/upform] save image %v.", result)
			}
			result = append(result[:len(result)-2], ']')
			io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.Bytes2str(result)+"}")
		} else {
			io.WriteString(resp, "{\"stat\": \"ioerr\"}")
			log.Errorln("[/upform] parse multipart form error.")
		}
	} else {
		io.WriteString(resp, "{\"stat\": \"noid\"}")
		log.Errorln("[/upform] no such user.")
	}
}

func vote(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if req.Method != "GET" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	q := req.URL.Query()
	uid, ok := q["uuid"]
	img, ok1 := q["img"]
	cls, ok2 := q["class"]
	if !ok || !ok1 || !ok2 {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		log.Errorln("[/vote] bad request.")
	} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
		if len([]rune(img[0])) == 5 {
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

func pickof(resp http.ResponseWriter, req *http.Request, isdl bool) {
	// 检查是否GET请求
	if req.Method != "GET" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	// 检查uid
	q := req.URL.Query()
	uid, ok := q["uuid"]
	if !ok {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		log.Errorln("[pickof] bad request.")
	} else if len([]rune(uid[0])) == 2 && userexists(uid[0]) {
		exclude := getkeys(conf.Users[uid[0]].Data)
		name := imago.Pick(exclude)
		if name == "" {
			io.WriteString(resp, "{\"stat\": \"nomoreimg\"}")
			log.Errorln("[pickof] no more image.")
		} else if isdl {
			imgpath := imgdir + name + ".webp"
			http.ServeFile(resp, req, imgpath)
			log.Println("[/pickdl]", name, ".")
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

func pick(resp http.ResponseWriter, req *http.Request) {
	pickof(resp, req, false)
}

func pickdl(resp http.ResponseWriter, req *http.Request) {
	pickof(resp, req, true)
}

func dice(resp http.ResponseWriter, req *http.Request) {
	// 检查是否GET请求
	if req.Method != "GET" {
		http.Error(resp, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var loli, noimg, newcls, r18, nopredict bool
	var link string
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
			http.ServeFile(resp, req, f)
		}
	} else {
		http.Error(resp, "500 Internal Server Error", http.StatusInternalServerError)
		log.Errorln("[/dice] predict error:", c, ".")
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
	if arglen == 6 || arglen == 7 {
		configfile = os.Args[2]
		imgdir = os.Args[3]
		custimgdir = os.Args[4]
		err := loadconf(configfile)
		if err != nil {
			panic(err)
		}
		go flushconf()
		if imgdir[len(imgdir)-1] != '/' {
			imgdir += "/"
		}
		if custimgdir[len(custimgdir)-1] != '/' {
			custimgdir += "/"
		}
		err = imago.Scanimgs(imgdir)
		if err != nil {
			panic(err)
		}
		pwd, _ = u82int(os.Args[5])
		pwdstr := (*[2]uintptr)(unsafe.Pointer(&os.Args[5]))
		for i := 0; i < len(os.Args[5]); i++ {
			*(*uint8)(unsafe.Pointer((*pwdstr)[0] + uintptr(i))) = '*'
		}
		listener, err := net.Listen("tcp", os.Args[1])
		if err != nil {
			panic(err)
		} else {
			if arglen == 7 {
				uid, err1 := strconv.Atoi(os.Args[6])
				if err == nil {
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
			log.Fatal(http.Serve(listener, nil))
		}
	} else {
		fmt.Println("Usage: <listen_addr> <configfile> <imgdir> <custimgdir> <password> (userid)")
	}
}
