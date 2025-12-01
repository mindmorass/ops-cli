package ui

import (
	"time"

	"github.com/briandowns/spinner"
)

// NewSpinner creates a new spinner with default settings
func NewSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + suffix
	return s
}

// StartSpinner starts a spinner and returns a stop function
func StartSpinner(suffix string) func() {
	s := NewSpinner(suffix)
	s.Start()
	return func() {
		s.Stop()
	}
}
