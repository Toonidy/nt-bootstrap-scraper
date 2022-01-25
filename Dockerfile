# Build Go API Server
FROM golang:1.17-buster AS go_builder
RUN go version
ARG BUILD_VERSION
ADD . /app
WORKDIR /app
RUN go build -o /main main.go

# Final stage build, this will be the container
# that we will deploy to production
FROM debian:buster

RUN useradd ntbootstrap
COPY --from=go_builder /main ./

RUN apt-get update; apt-get clean
RUN apt-get update && apt-get -y install wget
RUN wget --quiet https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt install -y ./google-chrome-stable_current_amd64.deb

# Execute Main Server
USER ntbootstrap
CMD ./main serve
