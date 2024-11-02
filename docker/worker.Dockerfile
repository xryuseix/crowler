FROM golang:1.23.2-alpine3.19
WORKDIR /app
RUN apk update && apk add make

COPY ../worker .

RUN go mod download
RUN go install github.com/cosmtrek/air@v1.49.0

CMD [ "air" ]