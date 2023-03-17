FROM golang:1.20-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY ./src /build
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go install
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main teamspeak.go
RUN chmod +x /build/main

FROM scratch
COPY --from=builder /build/main /app/
ENTRYPOINT ["/app/main"]