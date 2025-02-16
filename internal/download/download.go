package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/feline-dis/go-radio/internal/ingest"
)

type YtdlpResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Thumbnail   string `json:"thumbnail"`
	Duration    int    `json:"duration"`
	OriginalURL string `json:"original_url"`
	Filename    string `json:"_filename"`
}

type SongInfo struct {
	FileInfo os.FileInfo
	Duration int
}

type DownloadService struct {
	CachePath     string
	downloads     map[string]*SongInfo
	downloadQueue chan *ingest.Song
	numWorkers    int
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	activeJobs    sync.WaitGroup
}

func NewDownloadService(cachePath string, numWorkers int) *DownloadService {
	ctx, cancel := context.WithCancel(context.Background())
	ds := &DownloadService{
		CachePath:     cachePath,
		downloads:     make(map[string]*SongInfo),
		downloadQueue: make(chan *ingest.Song, 100),
		numWorkers:    numWorkers,
		ctx:           ctx,
		cancel:        cancel,
	}
	if err := ds.EnsureCacheDir(); err != nil {
		panic(fmt.Sprintf("failed to create cache directory: %v", err))
	}
	return ds
}

func (ds *DownloadService) EnsureCacheDir() error {
	if _, err := os.Stat(ds.CachePath); os.IsNotExist(err) {
		return os.MkdirAll(ds.CachePath, 0755)
	}
	return nil
}

func (ds *DownloadService) EnsureDownloaded(song *ingest.Song) error {
	if _, exists := ds.GetDownload(song.ID()); exists {
		return nil
	}

	if err := ds.QueueDownload(song); err != nil {
		fmt.Printf("Failed to queue download for %s: %v\n", song.URL, err)
		return err
	}

	return nil
}

func (ds *DownloadService) QueueDownload(song *ingest.Song) error {
	ds.activeJobs.Add(1) // Increment before queuing
	select {
	case ds.downloadQueue <- song:
		return nil
	case <-ds.ctx.Done():
		ds.activeJobs.Done() // Decrement if we couldn't queue
		return fmt.Errorf("download service is stopped")
	}
}

func (ds *DownloadService) GetDownload(id string) (*SongInfo, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	download, exists := ds.downloads[id]
	return download, exists
}

func (ds *DownloadService) Start() {
	for i := 0; i < ds.numWorkers; i++ {
		go ds.worker(i)
	}
}

func (ds *DownloadService) Stop() {
	ds.cancel()
}

func (ds *DownloadService) WaitForDownloads() {
	ds.activeJobs.Wait()
}

func (ds *DownloadService) worker(workerID int) {
	for {
		select {
		case song := <-ds.downloadQueue:
			if err := ds.downloadFile(song); err != nil {
				fmt.Printf("Worker %d failed to download %s: %v\n", workerID, song, err)
			}
			ds.activeJobs.Done() // Decrement when download is complete
		case <-ds.ctx.Done():
			fmt.Printf("Worker %d shutting down\n", workerID)
			return
		}
	}
}

func (ds *DownloadService) downloadFile(song *ingest.Song) error {
	if ds.downloads[song.ID()] != nil {
		return nil
	}
	// check if file already exists in filesystem by ID before downloading
	var ytdlpResponse *YtdlpResponse
	var fileInfo os.FileInfo
	if info, err := os.Stat(path.Join(ds.CachePath, song.ID()+".mp3")); err == nil {
		meta, err := os.ReadFile(path.Join(ds.CachePath, song.ID()+".json"))
		if err != nil {
			// file exists but metadata does not, download it
			ytdlpResponse, err = ds.downloadJSON(song.URL)
		}
		if err := json.Unmarshal(meta, &ytdlpResponse); err != nil {
			fmt.Println(string(meta))
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
		fileInfo = info
	} else {
		// file does not exist, download it
		ytdlpResponse, err = ds.downloadAudio(song.URL)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}

		fileInfo, err = os.Stat(path.Join(ds.CachePath, ytdlpResponse.ID+".mp3"))
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}
	}

	ds.saveMetadata(ytdlpResponse)

	ds.mu.Lock()
	ds.downloads[song.ID()] = &SongInfo{
		FileInfo: fileInfo,
		Duration: ytdlpResponse.Duration,
	}
	ds.mu.Unlock()

	return nil
}

func (ds *DownloadService) downloadAudio(url string) (*YtdlpResponse, error) {
	args := []string{
		"-x",
		"--audio-format",
		"mp3",
		"--print-json",
		"-o",
		ds.CachePath + "/%(id)s.%(ext)s",
		url,
	}

	cmd := exec.Command("yt-dlp", args...)

	stdout, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	return ds.parsdeYtdlpResponse(stdout)
}

func (ds *DownloadService) parsdeYtdlpResponse(stdout []byte) (*YtdlpResponse, error) {
	var ytdlpResponse YtdlpResponse

	if err := json.Unmarshal(stdout, &ytdlpResponse); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp response: %w", err)
	}

	return &ytdlpResponse, nil
}

func (ds *DownloadService) downloadJSON(url string) (*YtdlpResponse, error) {
	args := []string{
		"--skip-download",
		"-j",
		url,
	}

	cmd := exec.Command("yt-dlp", args...)
	fmt.Printf("yt-dlp %v\n", strings.Join(args, " "))
	stdout, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	return ds.parsdeYtdlpResponse(stdout)
}

func (ds *DownloadService) saveMetadata(ytdlpResponse *YtdlpResponse) error {
	json, err := json.Marshal(ytdlpResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	return os.WriteFile(path.Join(ds.CachePath, ytdlpResponse.ID+".json"), json, 0644)
}
