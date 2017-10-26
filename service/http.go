package service

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/apokalyptik/cshot/chrome"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

type Server struct {
	Worker      *chrome.Instance
	Chrome      string
	Host        string
	Port        int
	Concurrency int
}

func (s *Server) snap(url string) ([]byte, error) {
	buf, err := s.Worker.Screenshot(url)
	return buf, err

}

func (s *Server) ListenAndServe(procs int) error {
	worker, err := chrome.New(procs)
	if err != nil {
		return err
	}
	s.Worker = worker
	r := mux.NewRouter()
	r.PathPrefix("/cshot/v1/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pathParts = strings.SplitN(r.URL.Path, "/", 4)
		if len(pathParts) < 4 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var remoteURL = pathParts[3]
		buf, err := s.snap(remoteURL)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
		if wStr := r.URL.Query().Get("w"); wStr != "" {
			if width, err := strconv.Atoi(wStr); err == nil {
				src, _, err := image.Decode(bytes.NewReader(buf))
				if err == nil {
					dst := resize.Resize(uint(width), 0, src, resize.Lanczos3)
					w.Header().Set("Content-Type", "image/png")
					png.Encode(w, dst)
					return
				}
			}
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(buf)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		fmt.Fprintln(w, "/cshot/v1/{host}/{path}")
	})
	return http.ListenAndServe(
		fmt.Sprintf("%s:%d", s.Host, s.Port),
		handlers.ProxyHeaders(
			handlers.CombinedLoggingHandler(
				os.Stdout,
				handlers.CORS(
					handlers.AllowedOrigins([]string{"*"}),
					handlers.AllowedMethods([]string{"GET"}),
					handlers.AllowedHeaders([]string{"DPR"}),
				)(r),
			),
		),
	)
}
