package orchestrator

import (
	"fmt"
	"time"

	"github.com/feline-dis/go-radio/internal/controller"
	"github.com/feline-dis/go-radio/internal/download"
	"github.com/feline-dis/go-radio/internal/ingest"
	"github.com/feline-dis/go-radio/internal/picker"
)

type Orchestrator struct {
	downloadService     *download.DownloadService
	pickerService       *picker.PickerService
	websocketController *controller.WebsocketController

	current          *ingest.Song
	currentStartTime time.Time
	currentEndTime   time.Time
	next             *ingest.Song
}

func NewOrchestrator(downloadService *download.DownloadService, pickerService *picker.PickerService, wsc *controller.WebsocketController) *Orchestrator {
	return &Orchestrator{
		downloadService:     downloadService,
		pickerService:       pickerService,
		websocketController: wsc,
		current:             nil,
		currentStartTime:    time.Time{},
		currentEndTime:      time.Time{},
		next:                nil,
	}
}

func (o *Orchestrator) Start() {
	o.downloadService.Start()
	o.pickerService.ShuffleQueue()

	fmt.Println("Starting radio...")

	current := o.pickerService.NextSong()
	next := o.pickerService.NextSong()

	o.downloadService.QueueDownload(current)
	o.downloadService.QueueDownload(next)

	for {
		_, exists := o.downloadService.GetDownload(current.ID())

		if !exists {
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}

	info, _ := o.downloadService.GetDownload(current.ID())
	o.current = current
	o.currentStartTime = time.Now()
	o.currentEndTime = o.currentStartTime.Add(time.Duration(info.Duration) * time.Second)

	currentSongMessage := &controller.Message{
		Type: controller.MessageTypeCurrentSong,
		Payload: &controller.CurrentSongPayload{
			Title:     current.Title,
			Artist:    current.Artist,
			Duration:  info.Duration,
			ID:        current.ID(),
			StartTime: o.currentStartTime.Format(time.RFC3339),
			EndTime:   o.currentEndTime.Format(time.RFC3339),
		},
	}

	o.websocketController.BroadcastOnNewClient(currentSongMessage)
	o.websocketController.Broadcast(currentSongMessage)

	fmt.Printf("Playing %s\n", o.current.Title)

	for {
		time.Sleep(1 * time.Second)

		if time.Now().After(o.currentEndTime) {
			o.current = next
			o.currentStartTime = time.Now()
			nextInfo, exists := o.downloadService.GetDownload(next.ID())

			if !exists {
				continue
			}

			o.currentEndTime = o.currentStartTime.Add(time.Duration(nextInfo.Duration) * time.Second)

			next = o.pickerService.NextSong()
			o.downloadService.QueueDownload(next)

			fmt.Printf("Playing %s\n", o.current.Title)
			fmt.Printf("Next up: %s\n", next.Title)
		} else {
			fmt.Println("-----")
			fmt.Printf("Playing: %s\n", o.current.Title)
			fmt.Printf("Time elapsed: %v\n", time.Now().Sub(o.currentStartTime))
			fmt.Printf("Time remaining: %v\n", o.currentEndTime.Sub(time.Now()))
		}
	}
}
