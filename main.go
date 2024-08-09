package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/CodingVoid/gomble/gomble"
	"github.com/CodingVoid/gomble/logger"
)

// queue of tracks
var queue []*gomble.Track

func main() {
	loglevel, _ := strconv.Atoi(os.Getenv("GOMBLE_LOGLEVEL"))
	gomble.Init(logger.Loglevel(loglevel), os.Getenv("GOMBLE_SERVER")+":"+os.Getenv("GOMBLE_PORT"), false)
	gomble.Listener.OnPrivateMessageReceived = OnPrivateMessageReceived
	gomble.Listener.OnChannelMessageReceived = OnChannelMessageReceived
	gomble.Listener.OnTrackFinished = OnTrackFinished
	gomble.Listener.OnTrackPaused = OnTrackPaused
	gomble.Listener.OnTrackStopped = OnTrackStopped
	gomble.Listener.OnTrackException = OnTrackException

	gomble.Begin()
}

func OnPrivateMessageReceived(e gomble.PrivateMessageReceivedEvent) {
	gomble.SendMessageToUser("Send Back Private", e.Actor)
}

func OnChannelMessageReceived(e gomble.ChannelMessageReceivedEvent) {
	if strings.HasPrefix(e.Message, "#play ") {
		logger.Debugf(e.Message + "\n")
		tracklist, err := gomble.LoadUrl(e.Message)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
		queue = tracklist
		startNextTrack()
	} else if strings.HasPrefix(e.Message, "#play") {
		startNextTrack()
	} else if strings.HasPrefix(e.Message, "#stop") {
		gomble.Stop()
	} else if strings.HasPrefix(e.Message, "#pause") {
		gomble.Pause()
	} else if strings.HasPrefix(e.Message, "#resume") {
		gomble.Resume()
	} else if strings.HasPrefix(e.Message, "#add") {
		addToQueue(e.Message)
	} else if strings.HasPrefix(e.Message, "#next") {
		startNextTrack()
	} else if strings.HasPrefix(e.Message, "#list") {
		var list string
		for _, item := range queue {
			list = list + "<br /> " + item.GetTitle()
		}
		gomble.SendMessageToChannel("<br /><b>Queue</b>:<br />" + list, gomble.BotUserState.ChannelId)
	}
}

func OnTrackFinished(e gomble.TrackFinishedEvent) {
	startNextTrack()
}

func OnTrackPaused(e gomble.TrackPausedEvent) {
	logger.Infof("Paused Track: %s", e.Track.GetTitle())
}

func OnTrackStopped(e gomble.TrackStoppedEvent) {
	logger.Infof("Stopped Track: %s", e.Track.GetTitle())
}

func OnTrackException(e gomble.TrackExceptionEvent) {
	logger.Warnf("Got an Exception while playing Track: %s", e.Track.GetTitle())
}

func addToQueue(message string) {
	logger.Debugf(message + "\n")
		tracklist, err := gomble.LoadUrl(message)
		if err != nil {
			logger.Errorf("%v", err)
			return
		}
		queue = append(queue, tracklist...)
		// gomble.SendMessageToChannel("Added track <b>" + t.GetTitle() + "</b> to queue", gomble.BotUserState.ChannelId)
}

func startNextTrack() {
	if len(queue) > 0 {
		t := queue[0]
		// returns false if a track is already playing (or t == nil). returns true if starting was successful
		if gomble.Play(t) {
			gomble.SendMessageToChannel("Start playing Track <b>" + t.GetTitle() + "</b>", gomble.BotUserState.ChannelId)
			// If successful remove the track from the queue
			queue = queue[1:]
		}
	}
}
