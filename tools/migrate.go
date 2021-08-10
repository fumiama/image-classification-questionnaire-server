// Package main 接受两个参数usrdir imgdir
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/fumiama/image-classification-questionnaire-server/configo"
	"github.com/fumiama/imago"
	"github.com/sirupsen/logrus"
)

var conf configo.Data
var img2usr = make(map[string][]string)
var uploder = make(map[string]string)
var old2new = make(map[string]string)

func saveconf(configfile string) error {
	data, err := conf.Marshal()
	if err == nil {
		f, err1 := os.OpenFile(configfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 == nil {
			defer f.Close()
			_, err2 := f.Write(data)
			return err2
		}
		return err1
	}
	return err
}

func init() {
	imago.Setloglevel(logrus.ErrorLevel)
}

func main() {
	usrdir := os.Args[1]
	imgdir := os.Args[2]
	if imgdir[len(imgdir)-1] != '/' {
		imgdir += "/"
	}
	if usrdir[len(usrdir)-1] != '/' {
		usrdir += "/"
	}
	entry, err := os.ReadDir(usrdir)
	if err != nil {
		panic(err)
	}
	jsondata, err := os.ReadFile(imgdir + "info.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsondata, &uploder)
	if err != nil {
		panic(err)
	}
	fmt.Println("Unmashal success.")
	conf.Upload = make(map[string]string)
	conf.Users = make(map[string]*configo.DataVote)
	for _, i := range entry {
		if i.IsDir() {
			uid := i.Name()
			usrvote := usrdir + uid + "/"
			e, err := os.ReadDir(usrvote)
			if err != nil {
				panic(err)
			}
			for _, j := range e {
				if !j.IsDir() {
					name := j.Name()
					img2usr[name] = append(img2usr[name], uid)
					data, err := os.ReadFile(usrvote + name)
					if err != nil {
						panic(err)
					}
					class, err := strconv.Atoi(imago.Bytes2str(data))
					if err != nil {
						panic(err)
					}
					if conf.Users[uid] == nil {
						conf.Users[uid] = new(configo.DataVote)
						conf.Users[uid].Data = make(map[string]uint32)
					}
					conf.Users[uid].Data[name] = uint32(class)
				}
			}
		}
	}
	fmt.Println("Scan user dir success.")
	entry, err = os.ReadDir(imgdir)
	if err != nil {
		panic(err)
	}
	for _, i := range entry {
		if !i.IsDir() {
			name := i.Name()
			if strings.HasSuffix(name, ".webp") {
				dh := name[:len(name)-5]
				if len([]rune(dh)) == 5 {
					runtime.GC()
					f, err := os.ReadFile(imgdir + name)
					if err != nil {
						panic(err)
					}
					upnusi, ok := uploder[dh]
					_, newdh := imago.Saveimgbytes(f, imgdir, true, 1)
					if ok {
						conf.Upload[newdh] = upnusi
					}
					if newdh != dh {
						old2new[dh] = newdh
						uids, ok := img2usr[dh]
						if ok {
							for _, uid := range uids {
								class := conf.Users[uid].Data[dh]
								delete(conf.Users[uid].Data, dh)
								conf.Users[uid].Data[newdh] = class
							}
						}
						os.Remove(imgdir + name)
					}
				}
			}
		}
	}
	fmt.Println("Scan img dir success.")
	saveconf("conf.pb")
	fmt.Println("Save config success.")
	o2n, err := json.Marshal(&old2new)
	if err != nil {
		panic(err)
	}
	f, _ := os.Create("old2new.json")
	defer f.Close()
	f.Write(o2n)
	fmt.Println("Save old2new success.")
}
