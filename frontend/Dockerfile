FROM golang:1.22-bookworm AS builder

WORKDIR /src
COPY frontend/go.mod ./go.mod
COPY frontend/main.go ./main.go
RUN CGO_ENABLED=0 go build -o /markitdown-frontend .

FROM python:3.13-slim-bullseye

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

COPY --from=builder /markitdown-frontend /usr/local/bin/markitdown-frontend

EXPOSE 8080
ENTRYPOINT ["markitdown-frontend"]
