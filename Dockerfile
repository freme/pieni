FROM golang:alpine
ADD ./src /go/src/pieni
WORKDIR /go/src/pieni
RUN go get -d ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pieni .

FROM alpine
WORKDIR /root/
COPY --from=0 /go/src/pieni/pieni /bin/pieni
ADD ./src/static /root/static
#ENV PORT=3001
#EXPOSE 3001/tcp
CMD ["pieni"]

