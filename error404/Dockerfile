# thanks for the article at https://medium.com/@lizrice/non-privileged-containers-based-on-the-scratch-image-a80105d6d341
FROM ubuntu:latest as init
RUN useradd -u 10001 scratchuser

FROM golang:1.11 as build
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"'

FROM scratch as runtime
#FROM golang:1.11 as runtime
COPY --from=build /go/src/app/app /app/server
COPY ./config.json /app/
COPY ./media/ /app/media/
COPY ./templates/ /app/templates/
EXPOSE 8081
WORKDIR /app
ENTRYPOINT ["/app/server"]

COPY --from=init /etc/passwd /etc/passwd
USER scratchuser
