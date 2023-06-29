package cmd

import (
	"context"
	"flag"
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
	// "github.com/weaveworks/common/tracing"
)

var (
	// Define App host Url
	AppUrl = os.Getenv("APP_URL")
	app_str string
	c *client.Client
	logger log.Logger
	quit chan struct{}
	wg sync.WaitGroup
)

func init() {
	serverConfig := server.Config{
		MetricsNamespace: "demo",
	}
	serverConfig.RegisterFlags(flag.CommandLine)
	flag.Parse()
	serverConfig.LogLevel.Set("debug")

	logger = level.NewFilter(log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout)), serverConfig.LogLevel.Gokit)
	serverConfig.Log = logging.GoKit(logger)

	app, err := url.Parse(AppUrl)
	if err != nil {
		level.Error(logger).Log("msg", "<push_request init> error initializing tracing", "err", err)
		return
	}
	app_str = app.String()
	c = client.New(logger)
	quit = make(chan struct{})
}

func push_request(ctx context.Context, text string) {

	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(time.Duration(rand.Intn(2e3)) * time.Millisecond)
		for {
			select {
			case <-quit:
				return
			case <-timer.C:
				req, err := http.NewRequest("GET", app_str, nil)
				if err != nil {
					level.Error(logger).Log("msg", "<push_request timer> error building request", "err", err)
					return
				}
				req = req.WithContext(ctx)
				resp, err := c.Do(req)
				if err != nil {
					level.Error(logger).Log("msg", "<push_request timer> error doing request", "err", err)
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
				req, err := http.NewRequest("POST", app_str+"/post", strings.NewReader(form.Encode()))
				req = req.WithContext(ctx)
				if err != nil {
					level.Error(logger).Log("msg", "<push_request ticker> error building request", "err", err)
					return
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err := c.Do(req)
				if err != nil {
					level.Error(logger).Log("msg", "<push_request ticker> error doing request", "err", err)
					return
				}
				resp.Body.Close()
				return
			}
		}
	}()

	wg.Wait()
	return
}
