package errorskit

import (
	"fmt"
	"log"
)

// Wrap is a drop-in replacement for errors.Wrap (https://github.com/pkg/errors) using STD's fmt.Errorf().
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// LogWrap is a wrapper for log.Println and Wrap
func LogWrap(err error, message string) {
	log.Println(Wrap(err, message))
}

// FatalWrap is a wrapper for log.Fatalln and Wrap
func FatalWrap(err error, message string) {
	log.Fatalln(Wrap(err, message))
}
