FROM golang:1.14-buster
WORKDIR /usr/src/app
COPY go.mod go.sum /usr/src/app/
RUN go mod download
RUN go get github.com/githubnemo/CompileDaemon
COPY . /usr/src/app
RUN mkdir -p /usr/src/bin
run go build -o /usr/src/app/bin/ctfproxy /usr/src/app/main.go
EXPOSE 8000
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["gowatch"]

