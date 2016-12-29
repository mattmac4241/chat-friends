FROM golang:1.7
RUN mkdir -p /go/src/github.com/mattmac4241/chat-friends
WORKDIR /go/src/github.com/mattmac4241/chat-friends
COPY . /go/src/github.com/mattmac4241/chat-friends

ENV PORT 8081

EXPOSE 8081

CMD ["go", "run", "main.go"]
