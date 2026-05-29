FROM golang:1.23-bookworm@sha256:167053a2bb901972bf2c1611f8f52c44d5fe7e762e5cab213708d82c421614db AS builder

WORKDIR /src
COPY frontend/go.mod ./go.mod
COPY frontend/main.go ./main.go
RUN CGO_ENABLED=0 go build -o /markitdown-frontend .

FROM python:3.13-slim-trixie@sha256:b04b5d7233d2ad9c379e22ea8927cd1378cd15c60d4ef876c065b25ea8fb3bf3

ENV DEBIAN_FRONTEND=noninteractive
ENV EXIFTOOL_PATH=/usr/bin/exiftool
ENV FFMPEG_PATH=/usr/bin/ffmpeg

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    exiftool \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY packages/markitdown /app/packages/markitdown
RUN pip --no-cache-dir install /app/packages/markitdown[all]
RUN pip --no-cache-dir install yt-dlp

COPY --from=builder /markitdown-frontend /usr/local/bin/markitdown-frontend

EXPOSE 8080
ENTRYPOINT ["markitdown-frontend"]
