package service

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"

	"github.com/nfnt/resize"
)

var MaxWidth = 1280
var MinWidth = 20
var MaxHeight = 960
var MinHeight = 20

func (s *Server) output(w http.ResponseWriter, r *http.Request, buf []byte) {
	var setW uint
	var setH uint

	if hs := r.URL.Query().Get("h"); hs != "" {
		if height, _ := strconv.Atoi(hs); height > 0 {
			if height > MaxHeight {
				setH = uint(MaxHeight)
			} else if height < MinHeight {
				setH = uint(MinHeight)
			} else {
				setH = uint(height)
			}
		}
	}

	if ws := r.URL.Query().Get("w"); ws != "" {
		if width, _ := strconv.Atoi(ws); width > 0 {
			if width > MaxWidth {
				setW = uint(MaxWidth)
			} else if width < MinWidth {
				setW = uint(MinWidth)
			} else {
				setW = uint(width)
			}
		}
	}

	w.Header().Set("Content-Type", "image/png")
	if setW > 0 || setH > 0 {
		if src, _, err := image.Decode(bytes.NewReader(buf)); err != nil {
			log.Printf("Error decoding screenshot: %s", err.Error())
		} else {
			dst := resize.Resize(setW, setH, src, resize.Lanczos3)
			png.Encode(w, dst)
		}
	} else {
		w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
		w.Write(buf)
	}
}
