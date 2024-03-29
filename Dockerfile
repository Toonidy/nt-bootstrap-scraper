# Build Go API Server
FROM golang:1.17-buster AS go_builder
RUN go version
ARG BUILD_VERSION
ADD . /app
WORKDIR /app
RUN go build -o /main main.go

# Final stage build, this will be the container
# that we will deploy to production
FROM debian:bullseye

RUN useradd -m ntbootstrap
COPY --from=go_builder /main ./

RUN apt-get update && apt-get -y install wget gnupg
RUN echo "deb [arch=amd64] https://dl.google.com/linux/chrome/deb/ stable main" > /etc/apt/sources.list.d/google.list
RUN wget -O- https://dl.google.com/linux/linux_signing_key.pub |gpg --dearmor > /etc/apt/trusted.gpg.d/google.gpg
RUN apt-get update && apt-get install google-chrome-beta -y

# Execute Main Server
USER ntbootstrap
CMD ./main serve
