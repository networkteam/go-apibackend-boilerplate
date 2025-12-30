FROM alpine:3.14

RUN apk update &&\
    apk --no-cache add tzdata &&\
    apk --no-cache add ca-certificates &&\
    rm -rf /var/cache/apk/*

# Configure server timezone and PostgreSQL default timezone for sessions
ENV TZ=Europe/Berlin PGTZ=Europe/Berlin

WORKDIR /app

ENTRYPOINT ["./myproject-ctl"]

COPY ./bin/myproject-ctl ./myproject-ctl
