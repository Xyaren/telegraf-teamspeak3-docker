FROM golang:1.21-alpine as builder

WORKDIR /build

ADD ./go.mod ./go.sum ./
ADD ./teamspeak.go .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main teamspeak.go
RUN chmod +x /build/main

FROM scratch
COPY --from=builder /build/main /app/
ENTRYPOINT ["/app/main"]