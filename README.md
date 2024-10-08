[![Go Report Card](https://goreportcard.com/badge/github.com/CodingVoid/gomble)](https://goreportcard.com/report/github.com/CodingVoid/gomble)
# gomble
mumble library written in go. Intended for writing client side music bots.

## Using
- the main.go is intended as example of how the music bot could look like.
- Set server url, port and password via environment variables and then run gomble:
    ```
    GOMBLE_LOGLEVEL=1 \
    GOMBLE_SERVER=mumbleserverurl \
    GOMBLE_PORT=64738 \
    GOMBLE_PASSWORD=mumbleserverpassword \
    go run main.go
    ```
- If you don't want to study the entire Code in order to find out what you can do with this library and how, I made a README.md file in most folder explaining what each .go source file does. Furthermore the README file in the gomble directory shows a little illustration (sequence diagram) written in plantuml on how it works.

### Docker

A Dockerfile is included, if you want to run gomble via Docker:

1) Rename `.env.example` to `.env` and enter your mumble server url and password
2) Build Docker image with ```docker build -t gomble:latest .```
3) Run container with
    ```
    docker run ---env-file .env
    gomble:latest
    ```

### docker-compose
1) Rename `.env.example` to `.env` and enter your mumble server url and password
2) Build Docker image with ```docker build -t gomble:latest .```
3) Run container with
    ```
    docker-compose up -d
    ```

## Features
- you can play youtube videos ~~(without any additional dependency)~~ (for playing youtube-videos it is now necessary to have youtube-dl installed)
- it automatically uses UDP for sending audio data
- Buffering, so no disruptions in hearing "should" occur
- Stereo sound (since mumble 1.4 or newer)

## TODO
- implement more than just youtube videos as source for music
- be more or less OS independent (I am only using Linux and have not tested it on other Operating Systems)
- make library capable of using TLS certificates

## Notes
If you want to use this library be aware that this Project is still very much experimental. I appreciate and welcome any Issue or pull request or feature request.
If there are any questions, do not hesitate to write me an email (code.ivng5@simplelogin.co)

I got inspired by 'lavaplayer' (an audioplayer library for Discord) and 'gumble' (another mumble client implementation)
