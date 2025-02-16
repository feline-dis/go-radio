package picker

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/feline-dis/go-radio/internal/ingest"
)

type songWithPosition struct {
	song     *ingest.Song
	position float64
}

type PickerService struct {
	dataService *ingest.DataService
	AllSongs    []*ingest.Song
	queue       []*ingest.Song
	unpicked    []*ingest.Song
	quePos      int
}

func NewPickerService(ds *ingest.DataService) *PickerService {
	allSongs := SpotifyShuffle(ds.Songs)

	// Calculate queue size (2/3 of total songs)
	queueSize := 2 * (len(allSongs) / 3)

	// Create the service
	ps := &PickerService{
		AllSongs:    allSongs,
		dataService: ds,
		queue:       make([]*ingest.Song, queueSize),
		unpicked:    make([]*ingest.Song, len(allSongs)-queueSize),
		quePos:      0,
	}

	copy(ps.queue, allSongs[:queueSize])

	copy(ps.unpicked, allSongs[queueSize:])

	return ps
}

func (ps *PickerService) NextSong() *ingest.Song {
	fmt.Println("quePos: ", ps.quePos)
	fmt.Println("len(ps.queue): ", len(ps.queue))
	if ps.quePos >= len(ps.queue) {
		ps.ShuffleQueue()
		ps.quePos = 0
	}

	song := ps.queue[ps.quePos]
	ps.quePos++
	return song
}

func (ps *PickerService) ShuffleQueue() {
	fmt.Println("Shuffling queue...")

	// Combine current queue and unpicked songs
	allSongs := make([]*ingest.Song, len(ps.queue)+len(ps.unpicked))
	copy(allSongs, ps.queue)
	copy(allSongs[len(ps.queue):], ps.unpicked)

	// Shuffle all songs
	shuffled := SpotifyShuffle(allSongs)

	// Calculate new sizes (maintaining original ratio)
	totalSize := len(shuffled)
	queueSize := 2 * (totalSize / 3) // Same ratio as in SyncData

	// Create new queue and unpicked slices
	ps.queue = make([]*ingest.Song, queueSize)
	ps.unpicked = make([]*ingest.Song, totalSize-queueSize)

	// Distribute songs
	copy(ps.queue, shuffled[:queueSize])
	copy(ps.unpicked, shuffled[queueSize:])
}

func (ps *PickerService) SyncData() {
	ps.AllSongs = SpotifyShuffle(ps.dataService.Songs)

	queueSize := 2 * (len(ps.AllSongs) / 3)

	ps.queue = make([]*ingest.Song, queueSize)
	ps.unpicked = make([]*ingest.Song, len(ps.AllSongs)-queueSize)

	copy(ps.queue, SpotifyShuffle(ps.AllSongs[:queueSize]))
	copy(ps.unpicked, ps.AllSongs[queueSize:])

	ps.quePos = 0
}

func FisherYatesShuffle(list []*ingest.Song) {
	for i := len(list) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		(list)[i], (list)[j] = (list)[j], (list)[i]
	}
}

func SpotifyShuffle(songs []*ingest.Song) []*ingest.Song {
	// Step 1: Group songs by artist
	groups := make(map[string][]*ingest.Song)
	for _, song := range songs {
		artist := strings.ToLower(song.Artist)
		groups[artist] = append(groups[artist], song)
	}

	// Step 2: Shuffle each group and calculate positions
	var songsWithPositions []songWithPosition
	for _, group := range groups {
		// Shuffle the group
		FisherYatesShuffle(group)

		// Calculate group offset
		groupOffset := rand.Float64() * (1.0 / float64(len(group)))

		// Calculate positions for each song in group
		for idx, song := range group {
			// Calculate song offset similar to C# version
			songOffset := rand.Float64()*(0.2/float64(len(group))) -
				(0.1 / float64(len(group)))

			// Calculate final position
			position := float64(idx)/float64(len(group)) +
				groupOffset +
				songOffset

			songsWithPositions = append(songsWithPositions, songWithPosition{
				song:     song,
				position: position,
			})
		}
	}

	// Step 3: Sort by position
	sort.Slice(songsWithPositions, func(i, j int) bool {
		return songsWithPositions[i].position < songsWithPositions[j].position
	})

	// Step 4: Extract just the songs in their new order
	result := make([]*ingest.Song, len(songsWithPositions))
	for i, swp := range songsWithPositions {
		result[i] = swp.song
	}

	return result
}
