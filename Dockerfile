FROM alpine:3.10

RUN apk update &&\
    apk --no-cache add tzdata &&\
    apk --no-cache add ca-certificates &&\
    rm -rf /var/cache/apk/*

COPY ./docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
WORKDIR /

COPY ./docker/wait-for /wait-for
COPY persistence/migrations /migrations
COPY ./locales /locales
# This expects the binary to be built by a dependent CI job
COPY ./bin/myproject-ctl /myproject-ctl
