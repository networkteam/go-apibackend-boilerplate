module myvendor.mytld/myproject/backend

require (
	github.com/99designs/gqlgen v0.10.1
	github.com/99designs/gqlgen-contrib v0.0.0-20190913031219-de8886ed1b47
	github.com/apex/log v1.1.1
	github.com/friendsofgo/errors v0.8.1
	github.com/getsentry/sentry-go v0.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang-migrate/migrate/v4 v4.7.0
	github.com/gorilla/handlers v1.4.2
	github.com/lib/pq v1.2.0
	github.com/markbates/refresh v1.8.0
	github.com/mattn/go-isatty v0.0.9
	github.com/networkteam/go-sqllogger v0.2.0
	github.com/robfig/cron v1.2.0
	github.com/sethvargo/go-password v0.1.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.22.1
	github.com/vektah/gqlparser v1.1.2
	github.com/zbyte/go-kallax v1.3.9
	golang.org/x/crypto v0.0.0-20191202143827-86a70503ff7e
	gopkg.in/square/go-jose.v2 v2.4.0
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20190717161051-705d9623b7c1

go 1.13
