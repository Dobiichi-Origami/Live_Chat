FROM golang:1.17.2-alpine3.14
WORKDIR /runbin
COPY . ./src

RUN cd src &&\
    go env -w GO111MODULE=on &&\
    go mod tidy &&\
    go build -o ../liveChat liveChat/main &&\
    go clean -i -r -modcache -cache &&\
    cd ../ && rm -rf src

FROM alpine:latest
WORKDIR /root
COPY --from=0 /runbin/liveChat /root
ENTRYPOINT ./liveChat --path /appdata/config/config.json