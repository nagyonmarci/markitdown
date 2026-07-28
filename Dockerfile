FROM cgr.dev/chainguard/python:latest-dev@sha256:3be081f6cae8f1678609f6ae00b1dfebd6819c3ce75b7c574663af84afe99cc4

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
