/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context" // "log"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hirosassa/zerodriver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	telebot "gopkg.in/telebot.v3"
)

var (
	// TeleToken bot
	TeleToken = os.Getenv("TELE_TOKEN")
	// MetricsHost exporter host:port
	MetricsHost = os.Getenv("METRICS_HOST")
	// TracesHost exporter host:port
	TracesHost = os.Getenv("TRACES_HOST")
)
var otlp_grpc = "55680"

// Initialize OpenTelemetry
func initMetrics(ctx context.Context) {

	// Create a new OTLP Metric gRPC exporter with the specified endpoint and options
	exporter, _ := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(MetricsHost),
		otlpmetricgrpc.WithInsecure(),
	)

	// Define the resource with attributes that are common to all metrics.
	// labels/tags/resources that are common to all metrics.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(fmt.Sprintf("kbot_%s", AppVersion)),
	)

	// Create a new MeterProvider with the specified resource and reader
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			// collects and exports metric data every 10 seconds.
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second)),
		),
	)

	// Set the global MeterProvider to the newly created MeterProvider
	otel.SetMeterProvider(mp)

}

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initTraces(ctx context.Context) {

	logger := zerodriver.NewProductionLogger()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName("kbot-trace-service"),
			semconv.ServiceNameKey.String(AppVersion),
		),
	)
	if err != nil {
		logger.Fatal().Str("Error", err.Error()).Msg("<initTraces> failed to create resource: 'kbot-trace-service'")
		return
	}

	endpoint := strings.Split(TracesHost, ":")
	if len(endpoint) == 1 { TracesHost = TracesHost + ":" + otlp_grpc }

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(TracesHost), otlptracegrpc.WithInsecure())
	if err != nil {
		logger.Fatal().Str("Error", err.Error()).Msg("<initTraces> failed to create trace exporter")
		return
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

}

func push_metrics(ctx context.Context, payload string) {
	// Get the global MeterProvider and create a new Meter with the name "kbot_light_signal_counter"
	meter := otel.GetMeterProvider().Meter("kbot_command")

	// Get or create an Int64Counter instrument with the name "kbot_light_signal_<payload>"
	counter, _ := meter.Int64Counter(fmt.Sprintf("kbot_command_%s", payload))

	// Add a value of 1 to the Int64Counter
	counter.Add(ctx, 1)
}

// kbotCmd represents the kbot command
var kbotCmd = &cobra.Command{
	Use:     "kbot",
	Aliases: []string{"start"},
	Short:   "Start a bot",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		logger := zerodriver.NewProductionLogger()

		kbot, err := telebot.NewBot(telebot.Settings{
			URL:    "",
			Token:  TeleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		})

		if err != nil {
			logger.Fatal().Str("Error", err.Error()).Msg("Plaese check TELE_TOKEN env variable.")
			return
		} else {
			logger.Info().Str("Version", AppVersion).Msg("kbot started")
		}

		kbot.Handle(telebot.OnText, func(m telebot.Context) error {
			ctx := context.Background()
			tracer := otel.Tracer("kbot")
			ctx, span := tracer.Start(ctx,
				"OnText",
				trace.WithAttributes(attribute.String("component", "kbot")),
				trace.WithAttributes(attribute.String("TraceID", trace.TraceID{1, 2, 3, 4}.String())),
			)
			defer span.End()
			trace_id := span.SpanContext().TraceID().String()
			payload := m.Message().Payload
			msg_text := m.Text()
			msg_out := ""
			metric_label := "undefined"
			logger.Info().Str("TraceID", trace_id).Msg(payload)
			logger.Info().Str("Income message:", msg_text).Msg(payload)

			pushRequest := func(payload string) (string, string) {
				strTime := time.Now()
				push_request(payload)
				endTime := time.Now()
				duration := endTime.Sub(strTime)
				msg_out := fmt.Sprintf("Start request() at %s\nEnd request() at %s\nDuration: %s", strTime.Format("15:04:05.12340"), endTime.Format("15:04:05.12340"),duration)
				metric_label := "get"
				return msg_out, metric_label
			}
			switch payload {
			case "hello":
				err = m.Send(fmt.Sprintf("<b>Hello, %s</b>\nI'm %s!", m.Sender().FirstName, AppVersion), telebot.ModeHTML)
				metric_label = "hello"
			case "":
				switch msg_text {
				case "/start":
					err = m.Send("<b>Usage:</b>\n /help - for help message\n hello - to view 'hello message'\n ping - get 'Pong' response", telebot.ModeHTML)
					metric_label = "start"
				case "/help":
					err = m.Send("NP Kbot help page... be soon")
					metric_label = "help"
				case "/hello", "hello":
					err = m.Send(fmt.Sprintf("<b>Hello, %s</b>\nI'm %s!", m.Sender().FirstName, AppVersion), telebot.ModeHTML)
					metric_label = "hello"
				case "ping":
					err = m.Send("Pong")
					metric_label = "ping"
				case "/get":
					msg_out, metric_label = pushRequest(payload)
					err = m.Send(msg_out, telebot.ModeHTML)
				}
			default:
				switch msg_text {
				case "/get":
					msg_out, metric_label = pushRequest(payload)
					err = m.Send(msg_out, telebot.ModeHTML)
				default:
					err = m.Send("<b>Usage:</b>\n /help - for help message\n hello - to view 'hello message'\n ping - get 'Pong' response", telebot.ModeHTML)
				}
			}
			push_metrics(context.Background(), metric_label)
			return err
		})
		kbot.Start()
	},
}

func init() {
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	initMetrics(ctx)
	initTraces(ctx)
	rootCmd.AddCommand(kbotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// kbotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// kbotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
