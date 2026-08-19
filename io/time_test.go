package io_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	defer releaseSignal(release)
	finished := make(chan struct{})
	work := controlledResult(started, release, finished, "late", nil)

	result := runIO(io.WithTimeout[string](10 * time.Millisecond)(work))
	awaitSignal(t, started, "timed work to start")
	timedResult := awaitRunResult(t, result, "timeout result")
	require.ErrorIs(t, timedResult.Error, io.ErrorTimeout)

	releaseSignal(release)
	awaitSignal(t, finished, "timed work to finish after release")
}

func TestNotify(t *testing.T) {
	notification := make(chan io.GoResult[string], 1)
	ion := io.Notify(100*time.Millisecond, "a", func(str string, err error) {
		notification <- io.GoResult[string]{Value: str, Error: err}
	})

	runResult := awaitRunResult(t, runIO(ion), "Notify setup")
	require.NoError(t, runResult.Error)

	notificationResult := awaitRunResult(t, notification, "Notify callback")
	require.NoError(t, notificationResult.Error)
	require.Equal(t, "a", notificationResult.Value)
}
