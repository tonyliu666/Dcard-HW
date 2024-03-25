FROM golang:1.22.0 as builder
WORKDIR /go/src/github.com/tony/dcard-homework 
ADD . /go/src/github.com/tony/dcard-homework 
RUN go build -o dcard-homework
ENTRYPOINT ["./dcard-homework"]



