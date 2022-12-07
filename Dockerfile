FROM golang:alpine

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY isitwetbot.go .

RUN go install -v . 

CMD ["isitwetbot"]
