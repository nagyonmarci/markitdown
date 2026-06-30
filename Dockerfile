FROM cgr.dev/chainguard/python:latest-dev@sha256:90555ca52ffe2163ec7db49a8d8ee738f6e1c31e50729ee2c05cd4ad5f6ce043

ENV FFMPEG_PATH=/usr/bin/ffmpeg

USER root
RUN apk add --no-cache ffmpeg

WORKDIR /app
COPY . /app
RUN pip --no-cache-dir install \
    /app/packages/markitdown[all] \
    /app/packages/markitdown-sample-plugin

USER nonroot

ENTRYPOINT [ "markitdown" ]
