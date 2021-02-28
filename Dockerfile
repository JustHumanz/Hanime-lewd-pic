FROM golang:alpine

RUN apk update && apk add git
RUN echo "147.135.4.93 danbooru.donmai.us" >> /etc/hosts
COPY . /app
WORKDIR /app
RUN go build -o lewdpic
CMD ["./lewdpic"]