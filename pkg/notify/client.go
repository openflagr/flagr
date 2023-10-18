package notify

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/avast/retry-go"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

const contentTypeJSON = "application/json"
const userAgentHeader = "openflagr/flagr"

// CheckForRetry specifies a policy for handling retries
type CheckForRetry func(resp *http.Response, err error) (bool, error)

// DefaultRetryPolicy provides a default callback for Client.CheckForRetry
func DefaultRetryPolicy(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}

	if resp.StatusCode == 0 || resp.StatusCode >= 500 {
		return true, nil
	}

	return false, nil
}

// Backoff specifies a policy on how frequently we should retry
type Backoff func(min, max time.Duration, attemptNum int) time.Duration

// DefaultBackoff provides a default callback for Client.Backoff
func DefaultBackoff(min, max time.Duration, attemptNum int) time.Duration {
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}

var respReadLimit = int64(4096)

// Client holds a http.Client and retry configuration values
type Client struct {
	HTTPClient    *http.Client
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	AttemptsMax   int
	CheckForRetry CheckForRetry
	Backoff       Backoff
	Headers       http.Header
}

// NewClient spins up a Client with default retry configuration
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: config.Config.NotifyTimeout,
		},
		RetryWaitMin:  config.Config.NotifyRetryMin,
		RetryWaitMax:  config.Config.NotifyRetryMax,
		AttemptsMax:   config.Config.NotifyNumAttempts,
		CheckForRetry: DefaultRetryPolicy,
		Backoff:       DefaultBackoff,
	}
}

// Do executes a request
func (c *Client) Do(req *Request) (*http.Response, error) {
	i := 1
	var resp *http.Response
	err := retry.Do(
		func() error {
			if req.body != nil {
				if _, err := req.body.Seek(0, 0); err != nil {
					err = fmt.Errorf("failed to seek body: %v", err)
					return retry.Unrecoverable(err)
				}
			}

			// Attempt request
			resp, err := c.HTTPClient.Do(req.Request)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"reqMethod": req.Method,
					"reqURL":    req.URL,
				}).Warn("http request failed for notifier")
			}

			if err == nil {
				defer func() {
					c.drainBody(resp.Body)
				}()
			}

			// REVIEW: integrate with library retry.RetryIf
			checkOK, checkErr := c.CheckForRetry(resp, err)
			if !checkOK {
				if checkErr != nil {
					err = checkErr
				}

				return err
			}

			// REVIEW: integrate with library retry.OnRetry
			remain := c.AttemptsMax - i
			if remain == 0 {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"reqMethod": req.Method,
					"reqURL":    req.URL,
				}).Error("exhausted all attempts for request")

				return retry.Unrecoverable(err)
			}

			return err
		},
		retry.Attempts(uint(c.AttemptsMax)),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return c.Backoff(c.RetryWaitMin, c.RetryWaitMax, int(n))
		}),
	)

	return resp, err
}

func (c *Client) drainBody(body io.ReadCloser) {
	defer body.Close()
	_, err := io.Copy(io.Discard, io.LimitReader(body, respReadLimit))
	if err != nil {
		fmt.Printf("error reading response body: %v", err)
	}
}

// Request is the request with a replayable body
type Request struct {
	body io.ReadSeeker
	*http.Request
}

// NewRequest creates a request for a retryable request
func NewRequest(method, url string, body io.ReadSeeker) (*Request, error) {
	var rcBody io.ReadCloser
	if body != nil {
		rcBody = io.NopCloser(body)
	}

	httpReq, err := http.NewRequest(method, url, rcBody)
	if err != nil {
		return nil, err
	}

	return &Request{body, httpReq}, nil
}

// Post executes a POST request on a Retryable Client
func (c *Client) Post(url string, body io.ReadSeeker) (*http.Response, error) {
	req, err := NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("User-Agent", userAgentHeader)
	for k, v := range c.Headers {
		req.Header.Set(k, v[0])
	}
	return c.Do(req)
}
