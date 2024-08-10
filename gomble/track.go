package gomble

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/CodingVoid/gomble/gomble/audioformats"
	"github.com/CodingVoid/gomble/gomble/tracksources"
	"github.com/CodingVoid/gomble/gomble/tracksources/youtube"
	"github.com/CodingVoid/gomble/logger"
)

type Track struct {
	trackSrc tracksources.TrackSource
	// should be a multiple of audioformats.OPUS_FRAME_DURATION
	buffer_ms int
	// Never ever change that outside of this source file
	Done bool
}

var stop chan bool = make(chan bool)
var pause chan bool = make(chan bool)
var currTrack *Track

var enc, _ = audioformats.NewOpusEncoder() // initializes a new encoder

const (
	TRACK_TYPE_YOUTUBE = iota
	TRACK_TYPE_OGGFILE
)

var yttrackregex *regexp.Regexp = regexp.MustCompile(`https://(www\.)?youtu(be\.com|\.be)/(watch\?v=)?([a-zA-Z0-9\-\_]+)`) // need to use ` character otherwise \character are recognized as escape characters
var ytplaylistregex *regexp.Regexp = regexp.MustCompile(`https://(www\.)?youtube\.com/(playlist\?list=)([a-zA-Z0-9\-\_]+)`)

func LoadTrack(id string) (*Track, error) {
	var src tracksources.TrackSource
	if _, err := os.Stat("/usr/bin/yt-dlp"); err == nil {
		// use youtube-dl if it exists
		src, err = youtube.NewYoutubedlVideo(id)
		if err != nil {
			logger.Errorf("error creating video: %d", err)
			return nil, err
		}
	} else {
		// otherwise use native youtube stream implementation (probably doesn't work, but worth a try)
		// src, err = youtube.NewYoutubeVideo(surl)
		logger.Fatalf("No Youtube-dl installed. exiting\n")
	}

	return &Track{
		trackSrc: src,
	}, nil
}

func LoadPlaylist(id string) ([]*Track, error) {
	var idlist []string
	var tracklist []*Track

	// get all video ids
	cmd := exec.Command("yt-dlp", "--flat-playlist", "--print", "id", id)
	buf, err := cmd.Output()
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, fmt.Errorf("LoadPlaylist(%s:%d): %w", file, line, err)
	}

	// turn command results into slice of ids
	idstring := string(buf)
	idlist = strings.Split(idstring, "\n")
	// remove last element, because it is an empty string
	idlist = idlist[:len(idlist)-1]

	for _, id := range idlist {
		logger.Infof("ID: %v\n", id)
	}

	logger.Infof("Got playlist of length %v\n", strconv.Itoa(len(idlist)))

	for _, id := range idlist {
		track, err := LoadTrack(id)
		if err != nil {
			logger.Errorf("Could not load playlist track: %v\n", id)
			continue
		}
		logger.Infof("Loaded track: %v\n", track.trackSrc.GetTitle())
		tracklist = append(tracklist, track)
	}

	return tracklist, nil
}

func LoadUrl(url string) ([]*Track, error) {
	var tracklist []*Track

	ytplaylistmatches := ytplaylistregex.FindStringSubmatch(url)
	yttrackmatches := yttrackregex.FindStringSubmatch(url)

	if len(ytplaylistmatches) > 0 {
		// we have a playlist
		id := ytplaylistmatches[3]
		logger.Infof("LISTID: %v\n", id)
		tracklist, err := LoadPlaylist(id)
		if err != nil {
			return nil, err
		}
		return tracklist, nil
	} else if len(yttrackmatches) > 0 {
		// we have a single track
		id := yttrackmatches[4]
		track, err := LoadTrack(id)
		if err != nil {
			return nil, err
		}
		tracklist = append(tracklist, track)
		return tracklist, nil
	}
	return nil, errors.New("LoadTrack (track.go): No Youtube Video Found under URL")
}

func (t *Track) GetTitle() string {
	return t.trackSrc.GetTitle()
}

func GetCurrentTrack() *Track {
	return currTrack
}

func Play(t *Track) bool { // {{{
	if t.Done {
		logger.Debugf("Track is done")
		return false
	}
	if currTrack != nil {
		Stop()
		stop <- false
	}
	currTrack = t
	t.buffer_ms = 500 // buffering of 500 ms should be enough, I think...
	go audioroutine(t)
	return true
} // }}}

// Stops the current Track if one is playing
func Stop() {
	stop <- true
}

// Pauses the current Track if one is playing
func Pause() {
	pause <- true
}

func Resume() {
	pause <- false
}

// This audioroutine gets called whenever a new audio stream should be played
func audioroutine(t *Track) { // {{{
	// if err != nil {
	// 	eventpuffer <- TrackExceptionEvent{
	// 		Track: t,
	// 		err:   fmt.Errorf("Could not create Opus Encoder: %v\n", err),
	// 	}
	// 	return
	// }

	opusbuf := make(chan []byte, t.buffer_ms/audioformats.OPUS_FRAME_DURATION) // make channel with buffer_ms/frame_duration number of frames as buffer
	// our producer
	go func() {
		for {
			opusPayload, err := getNextOpusFrame(t, enc)
			if err != nil {
				logger.Errorf("Could not get next opus frame: %v\n", err)
				opusbuf <- nil
				return
			}
			if opusPayload == nil {
				// close channel instead of writing nil to channel (cleaner i think...)
				close(opusbuf)
				return
			}
			select {
			case opusbuf <- opusPayload:
				continue
			case <-stop:
				close(opusbuf)
				// stop <- false
				return
			}
		}
	}()

	timer := time.NewTicker((audioformats.OPUS_FRAME_DURATION) * time.Millisecond)
	var last bool
	//file, err := os.OpenFile("send.opus", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666)
	//if err != nil {
	//	logger.Fatal("...")
	//}
loop:
	for range timer.C {
		select {
		case <-stop:
			t.Done = true
			currTrack = nil
			eventpuffer <- TrackStoppedEvent{
				Track: t,
			}
			break loop
		case <-pause:
			eventpuffer <- TrackPausedEvent{
				Track: t,
			}
			select {
			case <-stop:
				t.Done = true
				currTrack = nil
				eventpuffer <- TrackStoppedEvent{
					Track: t,
				}
				break loop
			case <-pause:
				// Track resumed
				//TODO Check if pause is true or false (pause or resume)
			}
		default:
			logger.Debugf("Get next opus frame, number of frames in buffer: %d\n", len(opusbuf))
			lastTime := time.Now()
			opusPayload, ok := <-opusbuf
			elapsed := time.Since(lastTime)
			if elapsed.Microseconds() > 4000 {
				logger.Warnf("elapsed time from getting next opus frame: %s\n", elapsed)
			}
			if !ok {
				// Channel got closed -> track is done
				t.Done = true
				currTrack = nil
				eventpuffer <- TrackFinishedEvent{
					Track: t,
				}
				break loop
			}
			if opusPayload == nil {
				// happens if there was an error getting the opus payload from our producer
				t.Done = true
				currTrack = nil
				eventpuffer <- TrackExceptionEvent{
					Track: t,
				}
				break loop
			}
			sendAudioPacket(opusPayload, uint16(len(opusPayload)), last)

			//file.Write(opusPayload)
		}
		//WriteInt16InFile("pcm.final", allpcm)
	} // }}}
}

// var allpcm []int16
func getNextOpusFrame(t *Track, encoder *audioformats.OpusEncoder) ([]byte, error) { // {{{
	pcm, err := t.trackSrc.GetPCMFrame(audioformats.OPUS_FRAME_DURATION)
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, fmt.Errorf("Could not get PCM-Frame: %v", err)
	}
	opus, err := encoder.Encode(pcm)
	if err != nil {
		return nil, fmt.Errorf("Could not encode PCM-Frame: %v", err)
	}
	return opus, nil
} // }}}
