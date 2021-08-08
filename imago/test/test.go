package main

import (
	"fmt"
	"os"

	"github.com/fumiama/image-classification-questionnaire-server/imago"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/webp"
)

func main() {
	f, _ := os.Open("../../datas/imgs/滆愙畦玧擀.webp")
	img, _ := webp.Decode(f, &decoder.Options{})
	dh, _ := imago.GetDHashStr(img)
	fmt.Println(dh)
	fmt.Println(imago.HammDistance("滆愙畦玧擀", dh))
}
