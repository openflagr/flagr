package notify

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	"github.com/checkr/flagr/pkg/config"
	"github.com/sirupsen/logrus"
)

const contentTypeJSON = "application/json"
const userAgentHeader = "checkr/flagr"

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

var (
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4
	respReadLimit       = int64(4096)
)

// Client holds a http.Client and retry configuration values
type Client struct {
	HTTPClient    *http.Client
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	RetryMax      int
	CheckForRetry CheckForRetry
	Backoff       Backoff
}

// NewClient spins up a Client with default retry configuration
func NewClient() *Client {
	client := &Client{
		HTTPClient: &http.Client{
			Timeout: time.Duration(config.Config.NotifyTimeout * time.Second),
		},
		RetryWaitMin:  defaultRetryWaitMin,
		RetryWaitMax:  defaultRetryWaitMax,
		RetryMax:      defaultRetryMax,
		CheckForRetry: DefaultRetryPolicy,
		Backoff:       DefaultBackoff,
	}

	if config.Config.NotifyNumRetries != 0 {
		client.RetryMax = config.Config.NotifyNumRetries
	}
	if config.Config.RetryMin != 0 {
		client.RetryMax = config.Config.RetryMin
	}
	if config.Config.RetryMax != 0 {
		client.RetryMax = config.Config.RetryMax
	}

	return client
}

// Do executes a request
func (c *Client) Do(req *Request) (*http.Response, error) {
	i := 0

	for {
		var code int

		if req.body != nil {
			if _, err := req.body.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("failed to seek body: %v", err)
			}
		}

		// Attempt request
		resp, err := c.HTTPClient.Do(req.Request)
		checkOK, checkErr := c.CheckForRetry(resp, err)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":       err,
				"reqMethod": req.Method,
				"reqURL":    req.URL,
			}).Warn("http request failed for notifier")
		}

		if !checkOK {
			if checkErr != nil {
				err = checkErr
			}
			return resp, err
		}

		if err == nil {
			c.drainBody(resp.Body)
		}

		remain := c.RetryMax - i
		if remain == 0 {
			logrus.WithFields(logrus.Fields{
				"err":       err,
				"reqMethod": req.Method,
				"reqURL":    req.URL,
			}).Error("exhausted all attempts for request")
			break
		}
		wait := c.Backoff(c.RetryWaitMin, c.RetryWaitMax, i)

		desc := fmt.Sprintf("%s %s", req.Method, req.URL)
		if code > 0 {
			desc = fmt.Sprintf("%s (status: %d)", desc, code)
		}
		fmt.Println(desc)
		time.Sleep(wait)
		i++
	}

	return nil, fmt.Errorf("%s %s giving up after %d attempts", req.Method, req.URL, c.RetryMax+1)
}

func (c *Client) drainBody(body io.ReadCloser) {
	defer body.Close()
	_, err := io.Copy(ioutil.Discard, io.LimitReader(body, respReadLimit))
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
		rcBody = ioutil.NopCloser(body)
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
	return c.Do(req)
}
