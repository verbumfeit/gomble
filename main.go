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
		gomble.SendMessageToChannel("<b>Loading...</b>", gomble.BotUserState.ChannelId)

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
		for _, track := range queue {
			list = list + "<br /> " + track.GetTitle()
		}

		if gomble.GetCurrentTrack() != nil {
			gomble.SendMessageToChannel("<br /><br /><b>Now playing: </b>"+gomble.GetCurrentTrack().GetTitle()+"<br /><br /><b>Next up: </b><br />"+list+"<br />", gomble.BotUserState.ChannelId)
		} else if len(queue) > 0 {
			gomble.SendMessageToChannel("<br /><br /><b>Next up: </b><br />"+list+"<br />", gomble.BotUserState.ChannelId)
		} else {
			gomble.SendMessageToChannel("<br /><br /><b>Playlist is empty!</b><br /><br />", gomble.BotUserState.ChannelId)
		}
	} else if strings.HasPrefix(e.Message, "#help") {
		helpmessage := `
		<br />
		<br />

		I can play music from YouTube if you send a message like:<br />
		<br />

		<b>#play https://youtu.be/dQw4w9WgXcQ</b><br />
		<br />

		You can also <b>#pause</b> and <b>#resume</b> the music.<br />
		You can also <b>#add https://another-youtube.link</b> more songs to the playlist.<br />
		<br />

		Have fun!
		<br />
		`
		gomble.SendMessageToChannel(helpmessage, gomble.BotUserState.ChannelId)
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

	for _, track := range tracklist {
		gomble.SendMessageToChannel("Added track <b>"+track.GetTitle()+"</b> to queue", gomble.BotUserState.ChannelId)
	}
}

func startNextTrack() {
	if len(queue) > 0 {
		t := queue[0]
		// returns false if a track is already playing (or t == nil). returns true if starting was successful
		if gomble.Play(t) {
			gomble.SendMessageToChannel("Start playing Track <b>"+t.GetTitle()+"</b>", gomble.BotUserState.ChannelId)
			// If successful remove the track from the queue
			queue = queue[1:]
		}
	}
}
