// +build !windows

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/fumiama/imago"
	log "github.com/sirupsen/logrus"
)

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
