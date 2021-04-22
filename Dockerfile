ARG project_name=jfrog-cleaner
ARG project_version=1.0

FROM golang:1.14 as build
COPY ./ /app
WORKDIR /app
RUN make build-linux

FROM debian:9-slim

ENV NAME=app \
    UID=1001 \
    GID=1001 \
    TZ=Europe/Moscow
COPY --from=build app/jfrog-cleaner /usr/bin/
RUN groupadd --gid $GID $NAME && \
    useradd --uid $UID --gid $GID $NAME && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone
LABEL Name="${project_name}" Version="${project_version}"