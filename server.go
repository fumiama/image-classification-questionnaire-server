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
	var auth struct {
		Key int `http:"key"`
	}
	if err := unpack(req, &auth); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		fmt.Println("[/signup] bad request.")
	} else {
		diff := int(time.Now().Unix()) - (auth.Key ^ pwd)
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

func img(resp http.ResponseWriter, req *http.Request) {
	var data struct {
		Path string `http:"path"`
	}
	if err := unpack(req, &data); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		fmt.Println("[/img] bad request.")
	} else {
		if len([]rune(data.Path)) == 5 {
			imgpath := imgdir + data.Path + ".webp"
			if exists(imgpath) {
				http.ServeFile(resp, req, imgpath)
				fmt.Printf("[/img] serve %s.\n", data.Path)
			} else {
				resp.WriteHeader(404)
				io.WriteString(resp, "{\"stat\": \"nosuchimg\"}")
				fmt.Printf("[/img] %s not found.\n", data.Path)
			}
		} else {
			resp.WriteHeader(404)
			io.WriteString(resp, "{\"stat\": \"invimg\"}")
			fmt.Printf("[/img] invalid image path %s.\n", data.Path)
		}
	}
}

func upform(resp http.ResponseWriter, req *http.Request) {
	// 检查是否POST请求
	if req.Method != "POST" {
		resp.WriteHeader(405)
		return
	}
	// 检查uid
	var data struct {
		Uid string `http:"uuid"`
	}
	if err := unpack(req, &data); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		fmt.Println("[/upform] bad request.")
	} else if len([]rune(data.Uid)) == 2 && exists(usrdir+data.Uid) {
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
					result = append(result, imago.Str2bytes(saveimg(fo, data.Uid))...)
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
			log.Fatal(http.Serve(listener, nil))
		}
	} else {
		fmt.Println("Usage: <listen_addr> <usrdir> <imgdir> <password> (userid)")
	}
}
