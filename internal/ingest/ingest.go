package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/feline-dis/go-radio/internal/utils"
)

// DataService ingests songs and submitters from JSON files in the configured ingestPath.
type DataService struct {
	Submitters     map[string]Submitter
	submittersLock sync.RWMutex
	Songs          []*Song
	songsLock      sync.RWMutex
	ingestPath     string
	workQueue      chan string
	numWorkers     int
	ctx            context.Context
	cancel         context.CancelFunc
	activeJobs     sync.WaitGroup
}

// Submitter represents a submitter of songs.
type Submitter struct {
	Name string `json:"name"`
	Pfp  string `json:"pfp"`
}

// Song represents a song submitted by a submitter.
type Song struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
	ArtUrl string `json:"art_url"`
	URL    string `json:"url"`
}

// ID returns the YouTube video ID of the song or "" if an error is encountered.
func (s Song) ID() string {
	id, err := utils.ParseYouTubeVideoID(s.URL)

	if err != nil {
		return ""
	}

	return id
}

// SongList represents a list of songs submitted by a submitter.
type SongList struct {
	Name  string  `json:"name"`
	Pfp   string  `json:"pfp_url"`
	Songs []*Song `json:"songs"`
}

// NewDataService creates a new data service.
func NewDataService(ingestPath string, numWorkers int) *DataService {
	ctx, cancel := context.WithCancel(context.Background())
	return &DataService{
		Submitters: make(map[string]Submitter),
		Songs:      make([]*Song, 0),
		ingestPath: ingestPath,
		workQueue:  make(chan string, 100),
		numWorkers: numWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Ingest reads all JSON files in the ingest path and adds them to the data service.
func (ds *DataService) Ingest() error {
	files, err := os.ReadDir(ds.ingestPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Count valid files first
	validFiles := 0
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			validFiles++
		}
	}

	// Pre-increment WaitGroup for all valid files
	ds.activeJobs.Add(validFiles)

	// Queue all files
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(ds.ingestPath, file.Name())

		select {
		case ds.workQueue <- filePath:
			// Successfully queued
		case <-ds.ctx.Done():
			// Decrement WaitGroup for remaining files
			ds.activeJobs.Add(-validFiles) // Subtract all files we added
			return fmt.Errorf("data service is stopped")
		}
	}

	return nil
}

// Start starts the data service workers.
func (ds *DataService) Start() {
	for i := 0; i < ds.numWorkers; i++ {
		go ds.worker(i)
	}
}

// Stop stops the data service workers.
func (ds *DataService) Stop() {
	ds.cancel()
}

// WaitForJobs waits for all active jobs to complete.
func (ds *DataService) WaitForJobs() {
	ds.activeJobs.Wait()
}

func (ds *DataService) GetSongs() []*Song {
	ds.songsLock.RLock()
	defer ds.songsLock.RUnlock()
	return ds.Songs
}

func (ds *DataService) GetSong(id string) *Song {
	ds.songsLock.RLock()
	defer ds.songsLock.RUnlock()

	for _, song := range ds.Songs {
		if song.ID() == id {
			return song
		}
	}

	return nil
}

// ingestFile reads a JSON file and adds its contents to the data service.
func (ds *DataService) ingestFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var songList SongList
	if err := json.Unmarshal(data, &songList); err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Add songs to data service
	ds.songsLock.Lock()
	ds.Songs = append(ds.Songs, songList.Songs...)
	ds.songsLock.Unlock()

	// Add submitter to data service
	ds.submittersLock.Lock()
	ds.Submitters[songList.Name] = Submitter{
		Name: songList.Name,
		Pfp:  songList.Pfp,
	}
	ds.submittersLock.Unlock()

	return nil
}

// worker is a worker that processes files.
func (ds *DataService) worker(id int) {
	for {
		select {
		case filePath, ok := <-ds.workQueue:
			if !ok {
				// Channel is closed
				return
			}

			if err := ds.ingestFile(filePath); err != nil {
				fmt.Printf("Worker %d failed to process file %s: %v\n", id, filepath.Base(filePath), err)
			}
			ds.activeJobs.Done()

		case <-ds.ctx.Done():
			fmt.Printf("Worker %d shutting down\n", id)
			return
		}
	}
}
