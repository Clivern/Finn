FROM golang:1.20.4

ARG TIVO_VERSION=0.3.0

RUN mkdir -p /app/configs
RUN mkdir -p /app/var/logs
RUN apt-get update

WORKDIR /app

RUN curl -sL https://github.com/Clivern/Tivo/releases/download/v${TIVO_VERSION}/tivo_Linux_x86_64.tar.gz | tar xz
RUN rm LICENSE
RUN rm README.md

COPY ./config.dist.yml /app/configs/

EXPOSE 8080

VOLUME /app/configs
VOLUME /app/var

RUN ./tivo version

CMD ["./tivo", "server", "-c", "/app/configs/config.dist.yml"]
