package metrics

import (
	"math"
	"strconv"
	"time"

	"github.com/alexchebotarsky/thermostat-api/model/thermostat"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsHandled = newCollector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_handled",
		Help: "Handled requests counter and metadata associated with them",
	},
		[]string{"route_name", "status_code", "device_id"},
	))
	requestsDuration = newCollector(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "requests_duration",
		Help:    "Time spent processing requests",
		Buckets: []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1.0, 2.5, 5.0, 7.5, 10.0, math.Inf(1)},
	}, []string{"route_name", "status_code", "device_id"}))

	eventsProcessed = newCollector(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "events_processed",
		Help: "Handled PubSub events counter and metadata associated with them",
	},
		[]string{"event_name", "status", "device_id"},
	))
	eventsDuration = newCollector(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "events_duration",
		Help:    "Time spent processing events",
		Buckets: []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1.0, 2.5, 5.0, 7.5, 10.0, math.Inf(1)},
	},
		[]string{"event_name", "status", "device_id"},
	))

	thermostatMode = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermostat_mode",
		Help: "Mode of the thermostat",
	},
		[]string{"device_id"}))
	thermostatTargetTemperature = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermostat_target_temperature",
		Help: "Target temperature of the thermostat",
	},
		[]string{"device_id"}))
	thermostatOperatingState = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermostat_operating_state",
		Help: "Operating state of the thermostat",
	}, []string{"device_id"}))
	thermostatCurrentTemperature = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermostat_current_temperature",
		Help: "Current temperature reading of the thermostat",
	}, []string{"device_id"}))
	thermostatCurrentHumidity = newCollector(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "thermostat_current_humidity",
		Help: "Current humidity reading of the thermostat",
	}, []string{"device_id"}))
)

func AddRequestHandled(routeName string, statusCode int, deviceID string) {
	requestsHandled.WithLabelValues(routeName, strconv.Itoa(statusCode), deviceID).Inc()
}

func ObserveRequestDuration(routeName string, statusCode int, deviceID string, duration time.Duration) {
	requestsDuration.WithLabelValues(routeName, strconv.Itoa(statusCode), deviceID).Observe(duration.Seconds())
}

func AddEventProcessed(eventName, status, deviceID string) {
	eventsProcessed.WithLabelValues(eventName, status, deviceID).Inc()
}

func ObserveEventDuration(eventName, status, deviceID string, duration time.Duration) {
	eventsDuration.WithLabelValues(eventName, status, deviceID).Observe(duration.Seconds())
}

func SetThermostatMode(deviceID string, mode thermostat.Mode) {
	var modeValue float64
	switch mode {
	case thermostat.OffMode:
		modeValue = 0
	case thermostat.HeatMode:
		modeValue = 1
	case thermostat.CoolMode:
		modeValue = 2
	case thermostat.AutoMode:
		modeValue = 3
	default:
		modeValue = -1
	}

	thermostatMode.WithLabelValues(deviceID).Set(modeValue)
}

func SetThermostatTargetTemperature(deviceID string, temperature int) {
	thermostatTargetTemperature.WithLabelValues(deviceID).Set(float64(temperature))
}

func SetThermostatOperatingState(deviceID string, mode thermostat.OperatingState) {
	var modeValue float64
	switch mode {
	case thermostat.IdleOperatingState:
		modeValue = 0
	case thermostat.HeatingOperatingState:
		modeValue = 1
	case thermostat.CoolingOperatingState:
		modeValue = 2
	default:
		modeValue = -1
	}

	thermostatOperatingState.WithLabelValues(deviceID).Set(modeValue)
}

func SetThermostatCurrentTemperature(deviceID string, temperature float64) {
	thermostatCurrentTemperature.WithLabelValues(deviceID).Set(temperature)
}

func SetThermostatCurrentHumidity(deviceID string, humidity float64) {
	thermostatCurrentHumidity.WithLabelValues(deviceID).Set(humidity)
}
