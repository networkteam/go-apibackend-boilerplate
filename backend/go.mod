module myvendor.mytld/myproject/backend

require (
	github.com/99designs/gqlgen v0.17.49
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/apex/log v1.9.0
	github.com/boumenot/gocover-cobertura v1.2.0
	github.com/friendsofgo/errors v0.9.2
	github.com/getsentry/sentry-go v0.28.1
	github.com/go-jose/go-jose/v4 v4.0.3
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jackc/pgx/v5 v5.6.0
	github.com/joho/godotenv v1.5.1
	github.com/korylprince/go-graphql-ws v0.3.6
	github.com/mattn/go-isatty v0.0.20
	github.com/networkteam/apexlogutils v0.3.0
	github.com/networkteam/construct/v2 v2.0.1
	github.com/networkteam/qrb v0.8.0
	github.com/pressly/goose/v3 v3.21.1
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.9.0
	github.com/urfave/cli/v2 v2.27.2
	github.com/vektah/gqlparser/v2 v2.5.16
	github.com/wneessen/go-mail v0.4.2
	golang.org/x/crypto v0.25.0
	golang.org/x/term v0.22.0
)

// Bundled tool for JUnit xml reports of tests in CI
require gotest.tools/gotestsum v1.12.0

// Bundled tool for automatic build and restart of server during development
require github.com/networkteam/refresh v1.15.0

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/bitfield/gotestdox v0.2.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/dave/jennifer v1.7.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dnephin/pflag v1.0.7 // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/r3labs/sse/v2 v2.10.0 // indirect
	github.com/rjeczalik/notify v0.9.3 // indirect
	github.com/rs/cors v1.10.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/cobra v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.23.0 // indirect
	golang.org/x/xerrors v0.0.0-20240716161551-93cc26a95ae9 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// As noted on https://github.com/darccio/mergo this should fix an issue with a new vanity URL
replace github.com/imdario/mergo => github.com/imdario/mergo v0.3.16

go 1.22

toolchain go1.22.5
