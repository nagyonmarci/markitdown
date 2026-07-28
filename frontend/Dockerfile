FROM golang:1.26.5-bookworm@sha256:1ecb7edf62a0408027bd5729dfd6b1b8766e578e8df93995b225dfd0944eb651 AS builder

WORKDIR /src
COPY frontend/go.mod ./go.mod
COPY frontend/main.go ./main.go
RUN CGO_ENABLED=0 go build -o /markitdown-frontend .

FROM cgr.dev/chainguard/python:latest-dev@sha256:3be081f6cae8f1678609f6ae00b1dfebd6819c3ce75b7c574663af84afe99cc4

ENV FFMPEG_PATH=/usr/bin/ffmpeg

USER root
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY packages/markitdown /app/packages/markitdown
RUN pip --no-cache-dir install /app/packages/markitdown[all]
COPY packages/markitdown-ocr /app/packages/markitdown-ocr
RUN pip --no-cache-dir install /app/packages/markitdown-ocr[llm]
RUN pip --no-cache-dir install yt-dlp==2026.7.4

COPY --from=builder /markitdown-frontend /usr/local/bin/markitdown-frontend

USER nonroot
EXPOSE 8080
ENTRYPOINT ["markitdown-frontend"]
