FROM golang:latest AS compiled
RUN mkdir /program
COPY . /program
WORKDIR /program
RUN CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -a -o /go/bin/broker
FROM scratch
COPY --from=compiled /go/bin/broker /go/bin/broker
EXPOSE 8080
ENTRYPOINT ["/go/bin/broker"]