FROM golang:alpine as builder
RUN apk add --no-cache git
RUN mkdir /build
WORKDIR /build
#RUN git clone https://github.com/thannaske/telegraf-teamspeak3.git /build
RUN git clone https://github.com/Xyaren/telegraf-teamspeak3.git --branch patch-1 --single-branch /build
RUN go get -u github.com/thannaske/go-ts3
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM telegraf
ADD ./telegraf.conf /etc/telegraf/telegraf.conf
COPY --from=builder /build/main /app/
