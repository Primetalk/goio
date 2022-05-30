package io_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	start := time.Now()
	sleep1000ms := io.SleepA(1000*time.Millisecond, "a")
	atMost100ms := io.WithTimeout[string](100 * time.Millisecond)(sleep1000ms)
	_, err := io.UnsafeRunSync(atMost100ms)
	assert.Equal(t, io.ErrorTimeout, err)
	end := time.Now()
	assert.WithinDuration(t, end, start, 200*time.Millisecond)
}

func TestNotify(t *testing.T) {
	start := time.Now()
	notificationMoment := make(chan time.Time, 1)
	ion := io.Notify(100*time.Millisecond, "a", func(str string, err error) {
		assert.Equal(t, nil, err)

		notificationMoment <- time.Now()
	})
	_, err := io.UnsafeRunSync(ion)
	assert.Equal(t, nil, err)
	assert.WithinDuration(t, time.Now(), start, 10*time.Millisecond)
	time.Sleep(200 * time.Millisecond)
	assert.WithinDuration(t, <-notificationMoment, start, 200*time.Millisecond)
}
