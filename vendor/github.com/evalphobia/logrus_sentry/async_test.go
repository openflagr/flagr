package logrus_sentry

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestParallelLogging(t *testing.T) {
	WithTestDSN(t, func(dsn string, pch <-chan *resultPacket) {
		logger := getTestLogger()

		hook, err := NewAsyncSentryHook(dsn, []logrus.Level{
			logrus.ErrorLevel,
		})

		if err != nil {
			t.Fatal(err.Error())
		}
		logger.Hooks.Add(hook)

		wg := &sync.WaitGroup{}

		// start draining messages
		var logsReceived int
		const logCount = 10
		go func() {
			for i := 0; i < logCount; i++ {
				timeoutCh := time.After(hook.Timeout * 2)
				var packet *resultPacket
				select {
				case packet = <-pch:
				case <-timeoutCh:
					t.Fatalf("Waited %s without a response", hook.Timeout*2)
				}
				if packet.Logger != logger_name {
					t.Errorf("logger should have been %s, was %s", logger_name, packet.Logger)
				}

				if packet.ServerName != server_name {
					t.Errorf("server_name should have been %s, was %s", server_name, packet.ServerName)
				}
				logsReceived++
				wg.Done()
			}
		}()

		req, _ := http.NewRequest("GET", "url", nil)
		log := logger.WithFields(logrus.Fields{
			"server_name":  server_name,
			"logger":       logger_name,
			"http_request": req,
		})

		for i := 0; i < logCount; i++ {
			wg.Add(1)
			go func() {
				log.Error(message)
			}()
		}

		wg.Wait()
		if logCount != logsReceived {
			t.Errorf("Sent %d logs, received %d", logCount, logsReceived)
		}
	})
}
