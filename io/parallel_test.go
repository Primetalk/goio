package io_test

import (
	"testing"
	"time"

	"github.com/primetalk/goio/io"
	"github.com/stretchr/testify/assert"
)


func TestParallel(t *testing.T) {
	start := time.Now()
	sleep100ms := io.SleepA(100*time.Millisecond, "a")
	ios := []io.IO[string]{}
	for i := 0; i < 100; i+=1 {
		ios = append(ios, sleep100ms)
	}
	ioall := io.Parallel(ios)
	results, err := io.UnsafeRunSync(ioall)
	assert.Equal(t, err, nil)
	end := time.Now()
	assert.Equal(t, results[0], "a")
	assert.WithinDuration(t, end, start, 200*time.Millisecond)
}

func TestTimeout(t *testing.T) {
	start := time.Now()
	sleep1000ms := io.SleepA(1000*time.Millisecond, "a")
	atMost100ms := io.WithTimeout[string](100*time.Millisecond)(sleep1000ms)
	_, err := io.UnsafeRunSync(atMost100ms)
	assert.Equal(t, err, io.ErrorTimeout)
	end := time.Now()
	assert.WithinDuration(t, end, start, 200*time.Millisecond)
}
