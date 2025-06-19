# Dockerfile

FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o gojobberx

EXPOSE 8080

CMD ["./gojobberx"]
