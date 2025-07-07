FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS TARGETARCH TAGVERSION

WORKDIR /src
COPY --from=github . .
WORKDIR /build
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -C /src/ -o /build/ruuvi2db.app -ldflags "-X 'github.com/s5i/goutil/version.External=${TAGVERSION}'" .

FROM alpine
RUN apk add bash
COPY --from=build /src/entrypoint.sh /app/entrypoint.sh
COPY --from=build /src/example_config.yaml /app/example_config.yaml
COPY --from=build /build/ruuvi2db.app /app/ruuvi2db.app
VOLUME /appdata
CMD [ "/app/entrypoint.sh" ]
ENTRYPOINT /app/entrypoint.sh
