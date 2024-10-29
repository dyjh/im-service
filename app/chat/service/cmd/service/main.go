package main

import (
	"flag"
	"fmt"
	"im-service/app/chat/service/cmd/service/core"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	_ "go.uber.org/automaxprocs"
	"im-service/app/chat/service/internal/conf"
	"im-service/app/chat/service/utils"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "im-service_chat_service"
	// Version is the version of the compiled software.
	Version = "v0.0.1"
	// flagConf is the config flag.
	flagConf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, rr registry.Registrar) *kratos.App {

	IP, _ := utils.GetLocalIP()

	id = fmt.Sprintf("%s_chat_service", IP)

	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
		kratos.Registrar(rr),
	)
}

func main() {
	flag.Parse()
	/*logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)*/
	c := config.New(
		config.WithSource(
			file.NewSource(flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	var rc conf.Registry
	if err := c.Scan(&rc); err != nil {
		panic(err)
	}
	logLevel := strings.ToLower(bc.Log.Level)
	/*
		var loggerSetLevel log.Level
		switch logLevel {
		case "debug":
			loggerSetLevel = log.LevelDebug
			break
		case "info":
			loggerSetLevel = log.LevelInfo
			break
		case "warn":
			loggerSetLevel = log.LevelWarn
			break
		case "error":
			loggerSetLevel = log.LevelError
			break
		case "fatal":
			loggerSetLevel = log.LevelFatal
			break
		default:
			panic("日志配置错误")
		}*/

	// 创建 Zap Logger
	zapLogger := core.Zap(bc.Log)

	defer zapLogger.Sync()

	// 创建 Kratos 适配器的 Zap Logger
	kratosLogger := core.NewZapLoggerAdapter(zapLogger)

	// 设置服务的其他元数据
	logger := log.With(kratosLogger,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(bc.Trace.Endpoint)))
	if err != nil {
		panic(err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
		)),
	)

	app, h, cleanup, err := wireApp(&rc, &bc, logger, logLevel, tp)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	go h.ClientManager.WebSocketStart(bc.Websocket, logger)
	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
