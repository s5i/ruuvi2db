# docker build --platform linux/amd64,linux/arm64 -t shyym/ruuvi2db --push .

FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS TARGETARCH
RUN apk add --no-cache git

WORKDIR /git
RUN git clone https://github.com/s5i/ruuvi2db.git .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X 'github.com/s5i/goutil/version.External=$(git describe --tags)'" -o ruuvi2db.app .

FROM alpine

COPY --from=build /git/ruuvi2db.app /app/ruuvi2db.app

VOLUME /cfg
VOLUME /data

CMD [ "/app/ruuvi2db.app", "--config", "/cfg/ruuvi2db.cfg" ]
