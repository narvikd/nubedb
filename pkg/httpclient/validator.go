package httpclient

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func checkRes(errRes error) error {
	if errRes != nil {
		switch errRes.Error() {
		case "deadline exceeded":
			return errors.New("http client timed out")
		case "host is down":
			return errors.New("http client couldn't reach server. Server is down")
		case "connection refused":
			return errors.New("http client couldn't reach server. Connection refused")
		default:
			return fmt.Errorf("http client couldn't make http request: %w", errRes)
		}
	}
	return nil
}

func (r *Request) validateParams() error {
	if r.Timeout <= 0 {
		return errors.New("client timeout can't be 0 or lower")
	} else if r.Timeout > 6*time.Minute {
		return errors.New("client only accepts timeouts up to 6 minutes")
	}

	if r.Endpoint == "" {
		return errors.New("client endpoint is empty")
	}

	if !strings.Contains(r.Endpoint, "http") {
		return errors.New("client only accepts http/https URLs")
	}

	r.RestMethod = strings.ToUpper(r.RestMethod) // REST Methods must be upper case
	allowedRestMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for i := 0; i < len(allowedRestMethods); i++ {
		if r.RestMethod == allowedRestMethods[i] {
			break
		}

		if i == len(allowedRestMethods)-1 {
			return errors.New("method not recognized: " + r.RestMethod)
		}
	}

	return nil
}

// errContains is a helper function that returns if a lowered case str is contained in a lowered cased error
func errContains(err error, str string) bool {
	return strings.Contains(strings.ToLower(err.Error()), strings.ToLower(str))
}
