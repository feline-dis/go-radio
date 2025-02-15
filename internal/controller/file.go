package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/feline-dis/go-radio/internal/download"
	"github.com/feline-dis/go-radio/internal/ingest"
)

type FileController struct {
	r               *http.ServeMux
	downloadService *download.DownloadService
	dataService     *ingest.DataService
}

func NewFileController(r *http.ServeMux, downloadService *download.DownloadService, dataService *ingest.DataService) *FileController {
	return &FileController{
		r:               r,
		downloadService: downloadService,
		dataService:     dataService,
	}
}

func (fc *FileController) RegisterRoutes() {
	fc.r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs := http.FileServer(http.Dir("public"))

		fs.ServeHTTP(w, r)
	})
	fc.r.HandleFunc("/file/{id}", fc.getFile)
	fmt.Println("file routes registered")
}

func (fc *FileController) getFile(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/file/")
	fmt.Println("getting file", id)
	song := fc.dataService.GetSong(id)

	if song == nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	if err := fc.downloadService.EnsureDownloaded(song); err != nil {
		http.Error(w, "Failed to queue download", http.StatusInternalServerError)
		return
	}

	for {
		info, exists := fc.downloadService.GetDownload(song.ID())

		if !exists {
			fmt.Println("waiting for download")
			continue
		}

		http.ServeFile(w, r, fc.downloadService.CachePath+"/"+info.FileInfo.Name())
		break
	}
}
