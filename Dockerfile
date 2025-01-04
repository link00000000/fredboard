ARG FREDBOARD_SERVER_VERSION=""
ARG FREDBOARD_SERVER_COMMIT=""

FROM alpine:3.21 AS builder

RUN apk add opus opus-dev go make

RUN mkdir /build
COPY . /build
WORKDIR /build

RUN \
  FREDBOARD_SERVER_VERSION=${FREDBOARD_SERVER_VERSION} \
  FREDBOARD_SERVER_COMMIT=${FREDBOARD_SERVER_COMMIT} \
  make

FROM alpine:3.21

RUN apk add --no-cache opus go ffmpeg yt-dlp

RUN mkdir /app
COPY --from=builder /build/result/fredboard-server /app/fredboard-server

CMD ["/app/fredboard-server"]
