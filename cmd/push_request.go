package cmd

import (
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func push_request(text string) {
	serverConfig := server.Config{
		MetricsNamespace: "tns",
	}
	serverConfig.LogLevel.Set("debug")

	logger := level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), serverConfig.LogLevel.Gokit)
	serverConfig.Log = logging.GoKit(logger)

	os.Setenv("JAEGER_AGENT_HOST", TracesHost)
	os.Setenv("JAEGER_TAGS", "cluster=cloud,namespace=demo")
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")
	trace, err := tracing.NewFromEnv("kbot-j")
	if err != nil {
		level.Error(logger).Log("msg", "error initializing tracing", "err", err)
		return
	}
	defer trace.Close()

	app, err := url.Parse(AppUrl)
	if err != nil {
		level.Error(logger).Log("msg", "<push_request> error initializing tracing", "err", err)
		return
	}

	c := client.New(logger)
	quit := make(chan struct{})
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
			resp, err := c.Do(req)
			if err != nil {
				level.Error(logger).Log("msg", "error doing request", "err", err)
				return
			}
			resp.Body.Close()
			break
		}
	}

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			form := url.Values{}
			form.Add("text", text)
			req, err := http.NewRequest("POST", app.String()+"/post", strings.NewReader(form.Encode()))
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
			break
		}
	}
	close(quit)
	return
}
