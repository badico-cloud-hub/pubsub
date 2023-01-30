FROM golang:1.18

WORKDIR /app

COPY . ./
RUN go mod tidy

RUN go build -o /pubsub main.go

CMD [ "/pubsub" ]