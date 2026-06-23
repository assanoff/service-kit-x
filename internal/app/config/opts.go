// Package config defines the example application's configuration as a single
// go-flags options struct. Each field maps to a CLI flag (long), an environment
// variable (env, namespaced per group), a default, and a --help description.
package config

import "time"

// ServerOpts is the full application configuration. Groups are namespaced so
// their env vars read as HTTP_ADDR, GRPC_ADDR, DB_USER, OTEL_ENABLED, etc.
type ServerOpts struct {
	Service  string `long:"service-name" env:"SERVICE_NAME" default:"servicekit-example" description:"service name"`
	Env      string `long:"app-env"      env:"APP_ENV"      default:"development"         description:"runtime environment"`
	LogLevel string `long:"log-level"    env:"LOG_LEVEL"    default:"info"                description:"log level: debug|info|warn|error"`

	HTTP    HTTP    `group:"http" namespace:"http" env-namespace:"HTTP"`
	GRPC    GRPC    `group:"grpc" namespace:"grpc" env-namespace:"GRPC"`
	Debug   Debug   `group:"debug" namespace:"debug" env-namespace:"DEBUG"`
	DB      DB      `group:"postgres" namespace:"db" env-namespace:"DB"`
	OTEL    OTEL    `group:"otel" namespace:"otel" env-namespace:"OTEL"`
	Worker  Worker  `group:"worker" namespace:"worker" env-namespace:"WORKER"`
	Broker  Broker  `group:"broker" namespace:"broker" env-namespace:"BROKER"`
	Auth    Auth    `group:"auth" namespace:"auth" env-namespace:"AUTH"`
	Webhook Webhook `group:"webhook" namespace:"webhook" env-namespace:"WEBHOOK"`
}

// HTTP holds REST server settings. REST and gRPC run together or independently;
// each is on by default and can be turned off with its Disabled kill-switch
// (go-flags bools always default to false, so we model "default on" as Disabled).
type HTTP struct {
	Disabled          bool          `long:"disabled"            env:"DISABLED"            description:"disable the REST server"`
	Addr              string        `long:"addr"                env:"ADDR"                default:":8080"    description:"REST listen address"`
	ReadHeaderTimeout time.Duration `long:"read-header-timeout" env:"READ_HEADER_TIMEOUT" default:"5s"       description:"read header timeout"`
	RequestTimeout    time.Duration `long:"request-timeout"     env:"REQUEST_TIMEOUT"     default:"30s"      description:"per-request timeout"`
	ShutdownTimeout   time.Duration `long:"shutdown-timeout"    env:"SHUTDOWN_TIMEOUT"    default:"10s"      description:"graceful shutdown timeout"`
	BodySizeLimit     int64         `long:"body-size-limit"     env:"BODY_SIZE_LIMIT"     default:"1048576"  description:"max request body size in bytes"`
}

// GRPC holds gRPC server settings.
type GRPC struct {
	Disabled        bool          `long:"disabled"         env:"DISABLED"         description:"disable the gRPC server"`
	Addr            string        `long:"addr"             env:"ADDR"             default:":9090" description:"gRPC listen address"`
	ShutdownTimeout time.Duration `long:"shutdown-timeout" env:"SHUTDOWN_TIMEOUT" default:"10s"   description:"graceful shutdown timeout"`
	Reflection      bool          `long:"reflection"       env:"REFLECTION"       description:"enable server reflection"`
	MaxRecvMiB      int           `long:"max-recv-mib"     env:"MAX_RECV_MIB"     default:"16"    description:"max receive message size, MiB"`
	MaxSendMiB      int           `long:"max-send-mib"     env:"MAX_SEND_MIB"     default:"16"    description:"max send message size, MiB"`
}

// Debug toggles the debug endpoints (pprof + metrics + health). On by default
// like the other servers. WHERE they are served — attached to the application
// router or run on a separate internal port — is a wiring decision made in
// app/server (see server.go), not a runtime flag.
type Debug struct {
	Disabled bool   `long:"disabled" env:"DISABLED" description:"disable the debug/pprof endpoints"`
	Addr     string `long:"addr"     env:"ADDR"     default:"localhost:6060" description:"debug listen address (pprof, metrics, health)"`
}

// DB holds Postgres connection settings.
type DB struct {
	User         string `long:"user"           env:"USER"           default:"postgres"   description:"postgres user"            json:"-"`
	Password     string `long:"password"       env:"PASSWORD"       default:"postgres"   description:"postgres password"        json:"-"`
	Host         string `long:"host"           env:"HOST"           default:"localhost:5432" description:"postgres host:port"   json:"-"`
	Name         string `long:"name"           env:"NAME"           default:"servicekit" description:"postgres database name"   json:"-"`
	Schema       string `long:"schema"         env:"SCHEMA"         default:"public"     description:"postgres schema"          json:"-"`
	MaxIdleConns int    `long:"max-idle-conns" env:"MAX_IDLE_CONNS" default:"5"          description:"max idle connections"`
	MaxOpenConns int    `long:"max-open-conns" env:"MAX_OPEN_CONNS" default:"20"         description:"max open connections"`
	DisableTLS   bool   `long:"disable-tls"    env:"DISABLE_TLS"    description:"disable TLS to the database (set for local dev)"`
}

// Worker holds background-processing settings for the widget-import queue
// consumer. Like the servers, it is on by default and disabled via its
// kill-switch.
type Worker struct {
	Disabled      bool          `long:"disabled"       env:"DISABLED"       description:"disable background workers"`
	Interval      time.Duration `long:"interval"       env:"INTERVAL"       default:"1s" description:"queue poll interval"`
	BatchSize     int           `long:"batch-size"     env:"BATCH_SIZE"     default:"50" description:"max tasks claimed per tick"`
	CountInterval time.Duration `long:"count-interval" env:"COUNT_INTERVAL" default:"5s" description:"widget-count poller refresh interval"`
}

// Broker holds RabbitMQ + transactional-outbox settings. It is OFF by default
// (Enabled kill-switch inverted: brokers need external infra, unlike the
// servers/workers which default on). When disabled, widget.Create skips event
// publishing and no relay/consumer runs.
type Broker struct {
	Enabled    bool   `long:"enabled"     env:"ENABLED"     description:"enable RabbitMQ + outbox"`
	User       string `long:"user"        env:"USER"        default:"guest"            description:"rabbitmq user"        json:"-"`
	Password   string `long:"password"    env:"PASSWORD"    default:"guest"            description:"rabbitmq password"    json:"-"`
	Host       string `long:"host"        env:"HOST"        default:"localhost"        description:"rabbitmq host"`
	Port       string `long:"port"        env:"PORT"        default:"5672"             description:"rabbitmq port"`
	Source     string `long:"source"      env:"SOURCE"      default:"servicekit-example" description:"CloudEvents source"`
	Exchange   string `long:"exchange"    env:"EXCHANGE"    default:"servicekit.widgets" description:"widget events exchange"`
	RoutingKey string `long:"routing-key" env:"ROUTING_KEY" default:"widget.created"   description:"widget.created routing key"`
	Queue      string `long:"queue"       env:"QUEUE"       default:"servicekit.widget-audit" description:"consumer queue"`
}

// Auth holds JWT authentication settings for widget write endpoints. Off by
// default; when enabled a valid HMAC-signed JWT carrying RequiredRole is needed
// to create/update/delete widgets. Reads stay public.
type Auth struct {
	Enabled      bool   `long:"enabled"       env:"ENABLED"       description:"require JWT auth on widget writes"`
	JWTSecret    string `long:"jwt-secret"    env:"JWT_SECRET"    default:"" description:"HMAC secret for JWT verification" json:"-"`
	Issuer       string `long:"issuer"        env:"ISSUER"        default:"" description:"expected JWT issuer (optional)"`
	Audience     string `long:"audience"      env:"AUDIENCE"      default:"" description:"expected JWT audience (optional)"`
	RequiredRole string `long:"required-role" env:"REQUIRED_ROLE" default:"widget:write" description:"role required for widget writes"`
}

// Webhook holds the outbound widget.created webhook settings. OFF by default
// (it needs an external receiver). When enabled, an in-process eventbus consumer
// POSTs each created widget to URL using a resilient HTTP client that retries
// 429/503 with backoff (httpmw). It is independent of the broker/outbox.
type Webhook struct {
	Enabled     bool          `long:"enabled"      env:"ENABLED"      description:"POST widget.created to a webhook via the resilient HTTP client"`
	URL         string        `long:"url"          env:"URL"          default:""      description:"webhook endpoint receiving widget.created notifications"`
	Timeout     time.Duration `long:"timeout"      env:"TIMEOUT"      default:"3s"    description:"per-attempt timeout for webhook delivery"`
	MaxAttempts int           `long:"max-attempts" env:"MAX_ATTEMPTS" default:"4"     description:"max delivery attempts (retries on 429/503)"`
	BackoffBase time.Duration `long:"backoff-base" env:"BACKOFF_BASE" default:"200ms" description:"first retry delay"`
	BackoffMax  time.Duration `long:"backoff-max"  env:"BACKOFF_MAX"  default:"5s"    description:"max retry delay (cap)"`
}

// OTEL holds tracing settings.
type OTEL struct {
	Enabled     bool    `long:"enabled"     env:"ENABLED"     description:"enable OpenTelemetry tracing"`
	Endpoint    string  `long:"endpoint"    env:"ENDPOINT"    default:"localhost:4317" description:"OTLP/gRPC collector endpoint"`
	Insecure    bool    `long:"insecure"    env:"INSECURE"    description:"disable TLS to the collector"`
	Probability float64 `long:"probability" env:"PROBABILITY" default:"1.0"            description:"trace sampling probability"`
}
