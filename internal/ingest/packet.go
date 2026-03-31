package ingest

import (
	"encoding/json"
	"fmt"
	"time"
)

// Raw UDP message envelope — just enough to determine type and route.
type envelope struct {
	Type         string `json:"type"`
	SerialNumber string `json:"serial_number"`
	HubSN        string `json:"hub_sn"`
}

// Observation from obs_st message.
type Observation struct {
	Time               time.Time
	SerialNumber       string
	HubSN              string
	WindLull           *float64
	WindAvg            *float64
	WindGust           *float64
	WindDirection      *int16
	WindSampleInterval *int16
	StationPressure    *float64
	AirTemperature     *float64
	RelativeHumidity   *float64
	Illuminance        *int32
	UV                 *float64
	SolarRadiation     *int32
	RainAccumulated    *float64
	PrecipitationType  *int16
	LightningAvgDist   *int16
	LightningCount     *int16
	Battery            *float64
	ReportInterval     *int16
	FirmwareRevision   *int32
}

// RapidWind from rapid_wind message.
type RapidWind struct {
	Time          time.Time
	SerialNumber  string
	HubSN         string
	WindSpeed     float64
	WindDirection int16
}

// RainEvent from evt_precip message.
type RainEvent struct {
	Time         time.Time
	SerialNumber string
	HubSN        string
}

// LightningEvent from evt_strike message.
type LightningEvent struct {
	Time         time.Time
	SerialNumber string
	HubSN        string
	Distance     int16
	Energy       int32
}

// DeviceStatus from device_status message.
type DeviceStatus struct {
	Time             time.Time
	SerialNumber     string
	HubSN            string
	Uptime           int32
	Voltage          float64
	FirmwareRevision int32
	RSSI             int16
	HubRSSI          int16
	SensorStatus     int32
	Debug            int16
}

// HubStatus from hub_status message.
type HubStatus struct {
	Time             time.Time
	SerialNumber     string
	FirmwareRevision string
	Uptime           int32
	RSSI             int16
	ResetFlags       string
	Seq              int32
	RadioStatus      int16
	RadioRebootCount int32
	RadioI2CErrors   int32
	RadioNetworkID   int32
}

func epochToTime(v float64) time.Time {
	return time.Unix(int64(v), 0).UTC()
}

func optFloat(arr []any, i int) *float64 {
	if i >= len(arr) || arr[i] == nil {
		return nil
	}
	v, ok := arr[i].(float64)
	if !ok {
		return nil
	}
	return &v
}

func optInt16(arr []any, i int) *int16 {
	f := optFloat(arr, i)
	if f == nil {
		return nil
	}
	v := int16(*f)
	return &v
}

func optInt32(arr []any, i int) *int32 {
	f := optFloat(arr, i)
	if f == nil {
		return nil
	}
	v := int32(*f)
	return &v
}

func reqFloat(arr []any, i int) (float64, error) {
	if i >= len(arr) || arr[i] == nil {
		return 0, fmt.Errorf("missing field at index %d", i)
	}
	v, ok := arr[i].(float64)
	if !ok {
		return 0, fmt.Errorf("field at index %d is not a number", i)
	}
	return v, nil
}

// ParsePacket decodes a raw UDP packet into a typed struct.
// Returns the message type string and the parsed value, or an error.
func ParsePacket(data []byte) (string, any, error) {
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return "", nil, fmt.Errorf("invalid json: %w", err)
	}

	switch env.Type {
	case "obs_st":
		return parseObsST(data, env)
	case "rapid_wind":
		return parseRapidWind(data, env)
	case "evt_precip":
		return parseEvtPrecip(data, env)
	case "evt_strike":
		return parseEvtStrike(data, env)
	case "device_status":
		return parseDeviceStatus(data, env)
	case "hub_status":
		return parseHubStatus(data, env)
	default:
		return env.Type, nil, fmt.Errorf("unknown message type: %s", env.Type)
	}
}

func parseObsST(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Obs              [][]any `json:"obs"`
		FirmwareRevision *int32  `json:"firmware_revision"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "obs_st", nil, err
	}
	if len(raw.Obs) == 0 || len(raw.Obs[0]) < 18 {
		return "obs_st", nil, fmt.Errorf("obs_st: expected 18 fields, got %d", len(raw.Obs[0]))
	}

	arr := raw.Obs[0]
	epoch, err := reqFloat(arr, 0)
	if err != nil {
		return "obs_st", nil, fmt.Errorf("obs_st: %w", err)
	}

	obs := &Observation{
		Time:               epochToTime(epoch),
		SerialNumber:       env.SerialNumber,
		HubSN:              env.HubSN,
		WindLull:           optFloat(arr, 1),
		WindAvg:            optFloat(arr, 2),
		WindGust:           optFloat(arr, 3),
		WindDirection:      optInt16(arr, 4),
		WindSampleInterval: optInt16(arr, 5),
		StationPressure:    optFloat(arr, 6),
		AirTemperature:     optFloat(arr, 7),
		RelativeHumidity:   optFloat(arr, 8),
		Illuminance:        optInt32(arr, 9),
		UV:                 optFloat(arr, 10),
		SolarRadiation:     optInt32(arr, 11),
		RainAccumulated:    optFloat(arr, 12),
		PrecipitationType:  optInt16(arr, 13),
		LightningAvgDist:   optInt16(arr, 14),
		LightningCount:     optInt16(arr, 15),
		Battery:            optFloat(arr, 16),
		ReportInterval:     optInt16(arr, 17),
		FirmwareRevision:   raw.FirmwareRevision,
	}
	return "obs_st", obs, nil
}

func parseRapidWind(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Ob []any `json:"ob"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "rapid_wind", nil, err
	}
	if len(raw.Ob) < 3 {
		return "rapid_wind", nil, fmt.Errorf("rapid_wind: expected 3 fields, got %d", len(raw.Ob))
	}

	epoch, err := reqFloat(raw.Ob, 0)
	if err != nil {
		return "rapid_wind", nil, err
	}
	speed, err := reqFloat(raw.Ob, 1)
	if err != nil {
		return "rapid_wind", nil, err
	}
	dir, err := reqFloat(raw.Ob, 2)
	if err != nil {
		return "rapid_wind", nil, err
	}

	return "rapid_wind", &RapidWind{
		Time:          epochToTime(epoch),
		SerialNumber:  env.SerialNumber,
		HubSN:         env.HubSN,
		WindSpeed:     speed,
		WindDirection: int16(dir),
	}, nil
}

func parseEvtPrecip(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Evt []any `json:"evt"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "evt_precip", nil, err
	}
	if len(raw.Evt) < 1 {
		return "evt_precip", nil, fmt.Errorf("evt_precip: empty evt array")
	}
	epoch, err := reqFloat(raw.Evt, 0)
	if err != nil {
		return "evt_precip", nil, err
	}
	return "evt_precip", &RainEvent{
		Time:         epochToTime(epoch),
		SerialNumber: env.SerialNumber,
		HubSN:        env.HubSN,
	}, nil
}

func parseEvtStrike(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Evt []any `json:"evt"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "evt_strike", nil, err
	}
	if len(raw.Evt) < 3 {
		return "evt_strike", nil, fmt.Errorf("evt_strike: expected 3 fields, got %d", len(raw.Evt))
	}
	epoch, err := reqFloat(raw.Evt, 0)
	if err != nil {
		return "evt_strike", nil, err
	}
	dist, err := reqFloat(raw.Evt, 1)
	if err != nil {
		return "evt_strike", nil, err
	}
	energy, err := reqFloat(raw.Evt, 2)
	if err != nil {
		return "evt_strike", nil, err
	}
	return "evt_strike", &LightningEvent{
		Time:         epochToTime(epoch),
		SerialNumber: env.SerialNumber,
		HubSN:        env.HubSN,
		Distance:     int16(dist),
		Energy:       int32(energy),
	}, nil
}

func parseDeviceStatus(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Timestamp        float64 `json:"timestamp"`
		Uptime           float64 `json:"uptime"`
		Voltage          float64 `json:"voltage"`
		FirmwareRevision float64 `json:"firmware_revision"`
		RSSI             float64 `json:"rssi"`
		HubRSSI          float64 `json:"hub_rssi"`
		SensorStatus     float64 `json:"sensor_status"`
		Debug            float64 `json:"debug"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "device_status", nil, err
	}
	return "device_status", &DeviceStatus{
		Time:             epochToTime(raw.Timestamp),
		SerialNumber:     env.SerialNumber,
		HubSN:            env.HubSN,
		Uptime:           int32(raw.Uptime),
		Voltage:          raw.Voltage,
		FirmwareRevision: int32(raw.FirmwareRevision),
		RSSI:             int16(raw.RSSI),
		HubRSSI:          int16(raw.HubRSSI),
		SensorStatus:     int32(raw.SensorStatus),
		Debug:            int16(raw.Debug),
	}, nil
}

func parseHubStatus(data []byte, env envelope) (string, any, error) {
	var raw struct {
		Timestamp        float64 `json:"timestamp"`
		FirmwareRevision string  `json:"firmware_revision"`
		Uptime           float64 `json:"uptime"`
		RSSI             float64 `json:"rssi"`
		ResetFlags       string  `json:"reset_flags"`
		Seq              float64 `json:"seq"`
		RadioStats       []any   `json:"radio_stats"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "hub_status", nil, err
	}

	var radioStatus int16
	var radioRebootCount, radioI2CErrors, radioNetworkID int32
	if len(raw.RadioStats) >= 5 {
		radioRebootCount = int32(raw.RadioStats[1].(float64))
		radioI2CErrors = int32(raw.RadioStats[2].(float64))
		radioStatus = int16(raw.RadioStats[3].(float64))
		radioNetworkID = int32(raw.RadioStats[4].(float64))
	}

	return "hub_status", &HubStatus{
		Time:             epochToTime(raw.Timestamp),
		SerialNumber:     env.SerialNumber,
		FirmwareRevision: raw.FirmwareRevision,
		Uptime:           int32(raw.Uptime),
		RSSI:             int16(raw.RSSI),
		ResetFlags:       raw.ResetFlags,
		Seq:              int32(raw.Seq),
		RadioStatus:      radioStatus,
		RadioRebootCount: radioRebootCount,
		RadioI2CErrors:   radioI2CErrors,
		RadioNetworkID:   radioNetworkID,
	}, nil
}
