package metrics

import (
	"math"
	"strconv"
	"time"

	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsHandled = newCollector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_handled",
		Help: "Handled requests counter and metadata associated with them",
	},
		[]string{"route_name", "status_code"},
	))
	requestsDuration = newCollector(prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "requests_duration",
		Help:    "Time spent processing requests",
		Buckets: []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1.0, 2.5, 5.0, 7.5, 10.0, math.Inf(1)},
	}))

	eventsProcessed = newCollector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "events_processed",
		Help: "Handled PubSub events counter and metadata associated with them",
	},
		[]string{"event_name", "status"},
	))
	eventsDuration = newCollector(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "events_duration",
		Help:    "Time spent processing events",
		Buckets: []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1.0, 2.5, 5.0, 7.5, 10.0, math.Inf(1)},
	},
		[]string{"event_name"},
	))

	thermofridgeMode = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermofridge_mode",
		Help: "Mode of the thermofridge",
	},
		[]string{"device_id"}))
	thermofridgeTargetTemperature = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermofridge_target_temperature",
		Help: "Target temperature of the thermofridge",
	},
		[]string{"device_id"}))
	thermofridgeOperatingState = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermofridge_operating_state",
		Help: "Operating state of the thermofridge",
	}, []string{"device_id"}))
	thermofridgeCurrentTemperature = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermofridge_current_temperature",
		Help: "Current temperature reading of the thermofridge",
	}, []string{"device_id"}))
)

func AddRequestHandled(routeName string, statusCode int) {
	requestsHandled.WithLabelValues(routeName, strconv.Itoa(statusCode)).Inc()
}

func ObserveRequestDuration(duration time.Duration) {
	requestsDuration.Observe(duration.Seconds())
}

func AddEventProcessed(eventName, status string) {
	eventsProcessed.WithLabelValues(eventName, status).Inc()
}

func ObserveEventDuration(eventName string, duration time.Duration) {
	eventsDuration.WithLabelValues(eventName).Observe(duration.Seconds())
}

func SetThermofridgeMode(deviceID string, mode thermofridge.Mode) {
	var modeValue float64
	switch mode {
	case thermofridge.OffMode:
		modeValue = 0
	case thermofridge.HeatMode:
		modeValue = 1
	case thermofridge.CoolMode:
		modeValue = 2
	case thermofridge.AutoMode:
		modeValue = 3
	default:
		modeValue = -1
	}

	thermofridgeMode.WithLabelValues(deviceID).Set(modeValue)
}

func SetThermofridgeTargetTemperature(deviceID string, temperature int) {
	thermofridgeTargetTemperature.WithLabelValues(deviceID).Set(float64(temperature))
}

func SetThermofridgeOperatingState(deviceID string, mode thermofridge.OperatingState) {
	var modeValue float64
	switch mode {
	case thermofridge.IdleOperatingState:
		modeValue = 0
	case thermofridge.HeatingOperatingState:
		modeValue = 1
	case thermofridge.CoolingOperatingState:
		modeValue = 2
	default:
		modeValue = -1
	}

	thermofridgeOperatingState.WithLabelValues(deviceID).Set(modeValue)
}

func SetThermofridgeCurrentTemperature(deviceID string, temperature float64) {
	thermofridgeCurrentTemperature.WithLabelValues(deviceID).Set(temperature)
}
