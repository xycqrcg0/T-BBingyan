FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o bbingyan ./cmd

EXPOSE 8787

CMD ["/app/bbingyan"]

