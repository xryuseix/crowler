FROM golang:1.23.2-alpine3.19
WORKDIR /app
RUN apk update && apk add make

COPY ../app .

RUN go mod download

CMD [ "tail", "-f", "/dev/null" ]