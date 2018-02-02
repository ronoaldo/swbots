FROM golang:1.7.3 as builder
WORKDIR /go/src/github.com/ronoaldo/swbots/cmd/comlink
COPY *.go .
RUN go get -d -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/ronoaldo/swbots/cmd/comlink/comlink .
CMD ["./comlink"]
