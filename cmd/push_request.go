package cmd

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/tns/client"
	"github.com/weaveworks/common/logging"
	"github.com/weaveworks/common/server"
	"github.com/weaveworks/common/tracing"
)

var (
	// Define App host Url
	AppUrl = os.Getenv("APP_URL")
)

func push_request(ctx context.Context, text string) {
	serverConfig := server.Config{
		MetricsNamespace: "tns",
	}
	serverConfig.LogLevel.Set("debug")

	logger := level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), serverConfig.LogLevel.Gokit)
	serverConfig.Log = logging.GoKit(logger)

	/*  // duplicate metrics collector registration attempted [recovered]
	endpoint := strings.Split(TracesHost, ":")
	os.Setenv("JAEGER_AGENT_HOST", endpoint[0])
	os.Setenv("JAEGER_TAGS", "cluster=cloud,namespace=demo")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")
	trace, err := tracing.NewFromEnv("kbot-j")
	if err != nil {
		level.Error(logger).Log("msg", "error initializing tracing", "err", err)
		return
	}
	defer trace.Close()
	*/
	trace_id, exist := tracing.ExtractTraceID(ctx)
	if exist {
		level.Debug(logger).Log("msg", "<push_request> extract trace id:", "traceID", trace_id)
	} else {
		level.Debug(logger).Log("msg", "<push_request> can't extract trace id", "traceID", trace_id)
	}

	app, err := url.Parse(AppUrl)
	if err != nil {
		level.Error(logger).Log("msg", "<push_request> error initializing tracing", "err", err)
		return
	}

	c := client.New(logger)
	quit := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(time.Duration(rand.Intn(2e3)) * time.Millisecond)
		for {
			select {
			case <-quit:
				return
			case <-timer.C:
				req, err := http.NewRequest("GET", app.String(), nil)
				if err != nil {
					level.Error(logger).Log("msg", "error building request", "err", err)
					return
				}
				req = req.WithContext(ctx)
				resp, err := c.Do(req)
				if err != nil {
					level.Error(logger).Log("msg", "error doing request", "err", err)
					return
				}
				resp.Body.Close()
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-quit:
				return
			case <-ticker.C:
				form := url.Values{}
				form.Add("text", text)
				req, err := http.NewRequest("POST", app.String()+"/post", strings.NewReader(form.Encode()))
				req = req.WithContext(ctx)
				if err != nil {
					level.Error(logger).Log("msg", "error building request", "err", err)
					return
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err := c.Do(req)
				if err != nil {
					level.Error(logger).Log("msg", "error doing request", "err", err)
					return
				}
				resp.Body.Close()
				return
			}
		}
	}()
	// close(quit)
	wg.Wait()
	return
}
