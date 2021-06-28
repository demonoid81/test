FROM golang:1.15 as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
EXPOSE 8989
CMD [ "sh", "-c", "go run main.go server --tarantoolUrl=http://tarantool --db.host=roach"  ]