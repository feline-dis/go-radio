package orchestrator

import (
	"context"
	"fmt"
	"github.com/feline-dis/go-radio/internal/controller"
	"github.com/feline-dis/go-radio/internal/download"
	"github.com/feline-dis/go-radio/internal/ingest"
	"github.com/feline-dis/go-radio/internal/picker"
	"sync"
	"time"
)

type SongState struct {
	song      *ingest.Song
	startTime time.Time
	endTime   time.Time
	duration  int
}

type Orchestrator struct {
	downloadService     *download.DownloadService
	pickerService       *picker.PickerService
	websocketController *controller.WebsocketController
	current             *SongState
	next                *SongState
	mu                  sync.RWMutex
}

func NewOrchestrator(downloadService *download.DownloadService, pickerService *picker.PickerService, wsc *controller.WebsocketController) *Orchestrator {
	return &Orchestrator{
		downloadService:     downloadService,
		pickerService:       pickerService,
		websocketController: wsc,
	}
}

func (o *Orchestrator) Start() {
	ctx := context.Background()
	o.downloadService.Start()
	o.pickerService.ShuffleQueue()

	// Initialize first two songs
	if err := o.initializeFirstSongs(); err != nil {
		fmt.Printf("Failed to initialize first songs: %v\n", err)
		return
	}

	// Start the main playback loop
	go o.runPlaybackLoop(ctx)

	// Start the song transition timer
	go o.runTransitionTimer(ctx)
}

func (o *Orchestrator) initializeFirstSongs() error {
	// Get and prepare the first two songs
	firstSong := o.pickerService.NextSong()
	secondSong := o.pickerService.NextSong()

	// Queue both downloads
	o.downloadService.QueueDownload(firstSong)
	o.downloadService.QueueDownload(secondSong)

	// Wait for first song to be ready
	info, err := o.waitForDownload(firstSong.ID())
	if err != nil {
		return fmt.Errorf("failed to download first song: %w", err)
	}

	// Initialize the current song state
	now := time.Now()
	o.mu.Lock()
	o.current = &SongState{
		song:      firstSong,
		startTime: now,
		endTime:   now.Add(time.Duration(info.Duration) * time.Second),
		duration:  info.Duration,
	}
	o.next = &SongState{
		song: secondSong,
	}
	o.mu.Unlock()

	// Broadcast initial state
	o.broadcastCurrentSong()
	return nil
}

func (o *Orchestrator) waitForDownload(id string) (*download.SongInfo, error) {
	for attempts := 0; attempts < 60; attempts++ {
		if info, exists := o.downloadService.GetDownload(id); exists {
			return info, nil
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("timeout waiting for download of song %s", id)
}

func (o *Orchestrator) runPlaybackLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			o.mu.RLock()
			if time.Now().After(o.current.endTime) {
				o.mu.RUnlock()
				if err := o.transitionToNextSong(); err != nil {
					fmt.Printf("Error transitioning to next song: %v\n", err)
					time.Sleep(time.Second)
					continue
				}
			} else {
				fmt.Println("Playing song:", o.current.song.Title)
				fmt.Println("Next up:", o.next.song.Title)
				fmt.Println("Elapsed:", time.Since(o.current.startTime))
				o.mu.RUnlock()
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (o *Orchestrator) runTransitionTimer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			o.mu.RLock()
			timeUntilNext := time.Until(o.current.endTime)
			o.mu.RUnlock()

			if timeUntilNext > 0 {
				// Sleep until 1 second before the next transition
				time.Sleep(timeUntilNext - time.Second)
			}
		}
	}
}

func (o *Orchestrator) transitionToNextSong() error {
	// Ensure next song is downloaded
	nextInfo, err := o.waitForDownload(o.next.song.ID())
	if err != nil {
		return err
	}

	// Prepare the next-next song
	nextNextSong := o.pickerService.NextSong()
	o.downloadService.QueueDownload(nextNextSong)

	// Update state
	now := time.Now()
	o.mu.Lock()
	o.current = &SongState{
		song:      o.next.song,
		startTime: now,
		endTime:   now.Add(time.Duration(nextInfo.Duration) * time.Second),
		duration:  nextInfo.Duration,
	}
	o.next = &SongState{
		song: nextNextSong,
	}
	o.mu.Unlock()

	// Broadcast the change
	o.broadcastCurrentSong()
	return nil
}

func (o *Orchestrator) broadcastCurrentSong() {
	o.mu.RLock()
	defer o.mu.RUnlock()

	message := &controller.Message{
		Type: controller.MessageTypeCurrentSong,
		Payload: &controller.CurrentSongPayload{
			Title:     o.current.song.Title,
			Artist:    o.current.song.Artist,
			Duration:  o.current.duration,
			ID:        o.current.song.ID(),
			StartTime: o.current.startTime.Format(time.RFC3339),
			EndTime:   o.current.endTime.Format(time.RFC3339),
		},
	}

	o.websocketController.Broadcast(message)
	o.websocketController.BroadcastOnNewClient(message)
}

