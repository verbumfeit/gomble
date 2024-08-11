FROM golang:alpine


RUN apk add --no-cache go
RUN apk add --no-cache opus-dev
RUN apk add --no-cache yt-dlp

COPY . /gomble

WORKDIR /gomble

RUN go mod tidy

CMD [ "go", "run", "main.go" ]
