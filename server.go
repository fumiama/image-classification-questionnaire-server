package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/fumiama/image-classification-questionnaire-server/imago"
)

func getUUID() string {
	return "测试"
}

func index(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "index_quart.html")
}

var (
	pwd    int
	usrdir string
	imgdir string
)

func signup(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	key, ok := q["key"]
	if len(key) == 0 {
		http.Error(resp, "400 BAD REQUEST\nInvalid key.", http.StatusBadRequest)
		fmt.Println("[/signup] invalid key.")
	} else {
		keyint, err := strconv.Atoi(key[0])
		if !ok || err != nil {
			http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
			fmt.Println("[/signup] bad request.")
		} else {
			diff := int(time.Now().Unix()) - (keyint ^ pwd)
			if diff < 10 && diff >= 0 {
				uuid := getUUID()
				os.MkdirAll(usrdir+uuid, 0755)
				fmt.Printf("[/signup] create user: %s.\n", uuid)
				fmt.Fprintf(resp, "{\"stat\": \"success\", \"id\": \"%s\"}", url.QueryEscape((uuid)))
			} else {
				io.WriteString(resp, "{\"stat\": \"wrong\", \"id\": \"null\"}")
				fmt.Println("[/signup] auth failed.")
			}
		}
	}
}

func img(resp http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	path, ok := q["path"]
	if !ok || len(path) == 0 {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		fmt.Println("[/img] bad request.")
	} else {
		if len([]rune(path[0])) == 5 {
			imgpath := imgdir + path[0] + ".webp"
			if exists(imgpath) {
				http.ServeFile(resp, req, imgpath)
				fmt.Printf("[/img] serve %s.\n", path[0])
			} else {
				resp.WriteHeader(404)
				io.WriteString(resp, "{\"stat\": \"nosuchimg\"}")
				fmt.Printf("[/img] %s not found.\n", path[0])
			}
		} else {
			resp.WriteHeader(404)
			io.WriteString(resp, "{\"stat\": \"invimg\"}")
			fmt.Printf("[/img] invalid image path %s.\n", path[0])
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
	if !ok || len(uid) == 0 {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		fmt.Println("[/upload] bad request.")
	} else if len([]rune(uid[0])) == 2 && exists(usrdir+uid[0]) {
		if req.ContentLength <= 0 {
			io.WriteString(resp, "{\"stat\": \"emptybody\"}")
			fmt.Println("[/upload] invalid content length.")
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
			fmt.Printf("[/upload] receive %v/%v bytes.\n", cnt, req.ContentLength)
			if err == nil || cnt == cl {
				result = append(result, '{')
				result = append(result, imago.Str2bytes(saveimgbytes(buf, uid[0]))...)
				result = append(result, '}')
			} else {
				fmt.Printf("[/upload] receive body error: %v\n", err)
			}
			result = append(result, ']')
			io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.Bytes2str(result)+"}")
		}
	} else {
		io.WriteString(resp, "{\"stat\": \"noid\"}")
		fmt.Println("[/upload] no such user.")
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
	if !ok || len(uid) == 0 {
		http.Error(resp, "400 BAD REQUEST", http.StatusBadRequest)
		fmt.Println("[/upload] bad request.")
	} else if len([]rune(uid[0])) == 2 && exists(usrdir+uid[0]) {
		err := req.ParseMultipartForm(16 * 1024 * 1024)
		if err == nil {
			result := make([]byte, 1, 1024)
			result[0] = '['
			tail := imago.Str2bytes("}, ")
			for _, f := range req.MultipartForm.File["img"] {
				fmt.Printf("[/upform] receive %v of %v bytes.\n", f.Filename, f.Size)
				fo, err := f.Open()
				if err == nil {
					result = append(result, '{')
					result = append(result, imago.Str2bytes(saveimg(fo, uid[0]))...)
					result = append(result, tail...)
				} else {
					fmt.Printf("[/upform] save %v error.\n", f.Filename)
				}
			}
			result = append(result[:len(result)-2], ']')
			io.WriteString(resp, "{\"stat\": \"success\", \"result\": "+imago.Bytes2str(result)+"}")
		} else {
			io.WriteString(resp, "{\"stat\": \"ioerr\"}")
			fmt.Println("[/upform] parse multipart form error.")
		}
	} else {
		io.WriteString(resp, "{\"stat\": \"noid\"}")
		fmt.Println("[/upform] no such user.")
	}
}

func main() {
	arglen := len(os.Args)
	if arglen == 5 || arglen == 6 {
		usrdir = os.Args[2]
		imgdir = os.Args[3]
		if usrdir[len(usrdir)-1] != '/' {
			usrdir += "/"
		}
		if imgdir[len(imgdir)-1] != '/' {
			imgdir += "/"
		}
		pwd, _ = u82int(os.Args[4])
		pwdstr := (*[2]uintptr)(unsafe.Pointer(&os.Args[4]))
		for i := 0; i < len(os.Args[4]); i++ {
			*(*uint8)(unsafe.Pointer((*pwdstr)[0] + uintptr(i))) = '*'
		}
		listener, err := net.Listen("tcp", os.Args[1])
		if err != nil {
			panic(err)
		} else {
			if arglen == 6 {
				uid, err1 := strconv.Atoi(os.Args[5])
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
			// http.Handle("/yuka/", http.StripPrefix("/yuka/", http.FileServer(http.Dir(imgdir))))
			log.Fatal(http.Serve(listener, nil))
		}
	} else {
		fmt.Println("Usage: <listen_addr> <usrdir> <imgdir> <password> (userid)")
	}
}
