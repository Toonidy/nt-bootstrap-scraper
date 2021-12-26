# Build Go API Server
FROM golang:1.17-alpine AS go_builder
RUN go version
ADD . /app
WORKDIR /app
RUN go build -o /main main.go

# Final stage build, this will be the container
# that we will deploy to production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=go_builder /main ./

# Execute Main Server
RUN adduser -D nitrotype
USER nitrotype
CMD ./main serve --api_addr ":$PORT" --prod
