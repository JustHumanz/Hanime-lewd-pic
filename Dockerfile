FROM golang:alpine

RUN apk update && apk add git
COPY . /app
WORKDIR /app
RUN go build -o lewdpic
CMD ["./lewdpic"]