package api

import (
	"bytes"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
	"github.com/sphera-erp/sphera/internal/flow"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
)

var defaultFont *truetype.Font
var cjkFont *truetype.Font

func init() {
	fontBytes, err := ioutil.ReadFile("./fonts/font.ttf")
	if err != nil {
		fmt.Println(err)
	}
	defaultFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
	}
}

func QrcodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "image/png")
		writeImage(w, r)
	}
}

func writeImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code, err := vars["code"]
	if !err {
		http.Error(w, "Not found code", http.StatusNotFound)
		return
	}
	if content, ok := flow.JobStartRequestCodes[code]; ok {
		var pngqr []byte
		pngqr, err := qrcode.Encode(content, qrcode.Medium, 256)
		if err != nil {
			panic(err)
		}
		buffer := bytes.NewBuffer(pngqr)
		img, err := png.Decode(buffer)
		rect := img.Bounds()
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rect, img, rect.Min, draw.Src)
		addLabel(rgba, code)
		buff := bytes.NewBuffer([]byte{})
		png.Encode(buff, rgba)

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buff.Bytes())))
		if _, err := w.Write(buff.Bytes()); err != nil {
			fmt.Println(err)
			log.Println("unable to write image.")
		}
	} else {
		http.Error(w, "Not QR found", http.StatusNotFound)
		return
	}
}

func addLabel(img *image.RGBA, label string) {

	var fontSize = 24.0

	d := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: truetype.NewFace(defaultFont, &truetype.Options{
			Size: fontSize,
			DPI:  72,
		}),
	}
	d.Dot = fixed.Point26_6{
		X: (fixed.I(256) - d.MeasureString(label)) / 2,
		Y: fixed.I(220 + int(math.Ceil(fontSize * 1.35))),
	}
	d.DrawString(label)
}


