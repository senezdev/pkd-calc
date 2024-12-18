FROM golang:1.22.2-bullseye

RUN go install github.com/mitranim/gow@latest

WORKDIR /opt/app-root/src
COPY . /opt/app-root/src/

CMD [ "gow", "run", "pkd-bot/discord" ]
