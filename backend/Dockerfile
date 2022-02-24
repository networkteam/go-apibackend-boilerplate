FROM alpine:3.14

RUN apk update &&\
    apk --no-cache add tzdata &&\
    apk --no-cache add ca-certificates &&\
    rm -rf /var/cache/apk/*

# Configure server timezone and PostgreSQL default timezone for sessions
ENV TZ=Europe/Berlin PGTZ=Europe/Berlin

WORKDIR /app

COPY ./docker/entrypoint.sh ./entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]

COPY ./docker/wait-for ./wait-for

# TODO Use go:embed to include fixtures and migrations
COPY ./test/fixtures ./test/fixtures
COPY ./persistence/migrations ./persistence/migrations

COPY ./bin/myproject-ctl ./myproject-ctl
