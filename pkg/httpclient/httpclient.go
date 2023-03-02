package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Request struct {
	Endpoint   string
	RestMethod string
	BodyBytes  []byte `json:"-"` // '-' ignores the field on serialization
	Timeout    time.Duration
}

type Response struct {
	StatusCode StatusCode
	Err        error
	BodyBytes  []byte `json:"-"` // '-' ignores the field on serialization
}

type StatusCode struct {
	Code  int
	IsErr bool
}

func newResponse(statusCode int, err error, body []byte) Response {
	sc := StatusCode{
		Code:  statusCode,
		IsErr: true,
	}

	// Checks if it's in the ok range
	if sc.Code >= 200 && sc.Code <= 299 {
		sc.IsErr = false
	}

	res := Response{
		StatusCode: sc,
		Err:        err,
		BodyBytes:  body,
	}

	// If there isn't an error and the status code is not between the ok range. Set up a new error with it
	if res.Err == nil && res.StatusCode.IsErr {
		res.Err = errors.New("status code not ok: " + strconv.Itoa(res.StatusCode.Code))
	}

	return res
}

func (r *Request) Do(returnBody bool) Response {
	errVal := r.validateParams()
	if errVal != nil {
		return newResponse(http.StatusBadRequest, errVal, nil)
	}

	client, errClient := newClient(r.Timeout)
	if errClient != nil {
		return newResponse(http.StatusInternalServerError, errClient, nil)
	}

	req, errReq := http.NewRequest(r.RestMethod, r.Endpoint, bytes.NewReader(r.BodyBytes))
	if errReq != nil {
		return newResponse(http.StatusInternalServerError, errReq, nil)
	}

	res, errRes := client.Do(req)
	if checkRes(errRes) != nil {
		// Override some errors for convenience
		if errContains(errRes, "couldn't make http request") {
			return newResponse(http.StatusInternalServerError, errRes, nil)
		}

		if errContains(errRes, "no such host") || errContains(errRes, "connection refused") {
			return newResponse(http.StatusBadRequest, errors.New("host is down"), nil)
		}

		if errContains(errRes, "timeout") {
			return newResponse(http.StatusBadRequest, errors.New("timeout"), nil)
		}

		if errContains(errRes, "eof") {
			return newResponse(http.StatusBadRequest,
				errors.New("endpoint resetted while awaiting for an answer (EOF)"), nil)
		}

		return newResponse(http.StatusBadRequest, errRes, nil)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close() // Safe to ignore. GOSEC G307
	}(res.Body)

	if returnBody {
		resBody, errBody := getBody(res.Body)
		if errBody != nil {
			return newResponse(res.StatusCode, errBody, nil)
		}
		return newResponse(res.StatusCode, nil, resBody)
	}

	return newResponse(res.StatusCode, nil, nil)
}

// newClient returns a pointer to a new client which has all timeouts set to timeout,
//
// Exceptions:
//
// * Transport.ExpectContinueTimeout which is set to 1 second.
//
// * Transport.IdleConnTimeout set to 30 seconds.
func newClient(timeout time.Duration) (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, timeout)
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: timeout,
	}
	return client, nil
}

func getBody(body io.ReadCloser) ([]byte, error) {
	b, errReadBody := io.ReadAll(body)
	if errReadBody != nil {
		return nil, fmt.Errorf("couldn't read response body: %w", errReadBody)
	}

	if string(b) == "" {
		return nil, errors.New("body returned empty")
	}

	return b, nil
}
