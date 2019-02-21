// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package main

import (
	"fmt"

	"git.curoverse.com/arvados.git/sdk/go/httpserver"
	"github.com/prometheus/client_golang/prometheus"
)

type nodeMetrics struct {
	reg *prometheus.Registry
}

func (m *nodeMetrics) setupBufferPoolMetrics(b *bufferPool) {
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "bufferpool_bytes_allocated",
			Help:      "Number of bytes allocated to buffers",
		},
		func() float64 { return float64(b.Alloc()) },
	))
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "bufferpool_buffers_max",
			Help:      "Maximum number of buffers allowed",
		},
		func() float64 { return float64(b.Cap()) },
	))
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "bufferpool_buffers_in_use",
			Help:      "Number of buffers in use",
		},
		func() float64 { return float64(b.Len()) },
	))
}

func (m *nodeMetrics) setupWorkQueueMetrics(q *WorkQueue, qName string) {
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      fmt.Sprintf("%s_queue_in_progress", qName),
			Help:      fmt.Sprintf("Number of %s requests in progress", qName),
		},
		func() float64 { return float64(getWorkQueueStatus(q).InProgress) },
	))
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      fmt.Sprintf("%s_queue_queued", qName),
			Help:      fmt.Sprintf("Number of queued %s requests", qName),
		},
		func() float64 { return float64(getWorkQueueStatus(q).Queued) },
	))
}

func (m *nodeMetrics) setupRequestMetrics(rc httpserver.RequestCounter) {
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "requests_current",
			Help:      "Number of requests in progress",
		},
		func() float64 { return float64(rc.Current()) },
	))
	m.reg.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "requests_max",
			Help:      "Maximum number of concurrent requests",
		},
		func() float64 { return float64(rc.Max()) },
	))
}

type volumeMetricsVecs struct {
	ioBytes     *prometheus.CounterVec
	errCounters *prometheus.CounterVec
	opsCounters *prometheus.CounterVec
}

func newVolumeMetricsVecs(reg *prometheus.Registry) *volumeMetricsVecs {
	m := &volumeMetricsVecs{}
	m.opsCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "volume_operations",
			Help:      "Number of volume operations",
		},
		[]string{"device_id", "operation"},
	)
	reg.MustRegister(m.opsCounters)
	m.errCounters = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "volume_errors",
			Help:      "Number of volume errors",
		},
		[]string{"device_id", "error_type"},
	)
	reg.MustRegister(m.errCounters)
	m.ioBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "arvados",
			Subsystem: "keepstore",
			Name:      "volume_io_bytes",
			Help:      "Volume I/O traffic in bytes",
		},
		[]string{"device_id", "direction"},
	)
	reg.MustRegister(m.ioBytes)

	return m
}
