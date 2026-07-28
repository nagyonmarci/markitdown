FROM golang:1.26-bookworm@sha256:b305420a68d0f229d91eb3b3ed9e519fcf2cf5461da4bef997bf927e8c0bfd2b AS builder

WORKDIR /src
COPY frontend/go.mod ./go.mod
COPY frontend/main.go ./main.go
RUN CGO_ENABLED=0 go build -o /markitdown-frontend .

FROM cgr.dev/chainguard/python:latest-dev@sha256:90555ca52ffe2163ec7db49a8d8ee738f6e1c31e50729ee2c05cd4ad5f6ce043

ENV FFMPEG_PATH=/usr/bin/ffmpeg

USER root
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY packages/markitdown /app/packages/markitdown
RUN pip --no-cache-dir install /app/packages/markitdown[all]
COPY packages/markitdown-ocr /app/packages/markitdown-ocr
RUN pip --no-cache-dir install /app/packages/markitdown-ocr[llm]
RUN pip --no-cache-dir install yt-dlp==2026.6.9

COPY --from=builder /markitdown-frontend /usr/local/bin/markitdown-frontend

USER nonroot
EXPOSE 8080
ENTRYPOINT ["markitdown-frontend"]
