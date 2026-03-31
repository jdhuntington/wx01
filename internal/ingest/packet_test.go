package ingest

import (
	"testing"
	"time"
)

func TestParseObsST(t *testing.T) {
	data := []byte(`{
		"serial_number": "ST-00000512",
		"type": "obs_st",
		"hub_sn": "HB-00013030",
		"obs": [[1588948614, 0.18, 0.22, 0.27, 144, 6, 1017.57, 22.37, 50.26, 328, 0.03, 3, 0.000000, 0, 0, 0, 2.410, 1]],
		"firmware_revision": 129
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "obs_st" {
		t.Fatalf("expected obs_st, got %s", msgType)
	}

	obs := parsed.(*Observation)
	if obs.SerialNumber != "ST-00000512" {
		t.Errorf("serial_number = %q", obs.SerialNumber)
	}
	if obs.HubSN != "HB-00013030" {
		t.Errorf("hub_sn = %q", obs.HubSN)
	}
	if obs.Time != time.Unix(1588948614, 0).UTC() {
		t.Errorf("time = %v", obs.Time)
	}
	if *obs.AirTemperature != 22.37 {
		t.Errorf("air_temperature = %v", *obs.AirTemperature)
	}
	if *obs.StationPressure != 1017.57 {
		t.Errorf("station_pressure = %v", *obs.StationPressure)
	}
	if *obs.Battery != 2.41 {
		t.Errorf("battery = %v", *obs.Battery)
	}
	if *obs.FirmwareRevision != 129 {
		t.Errorf("firmware_revision = %v", *obs.FirmwareRevision)
	}
}

func TestParseObsSTWithNulls(t *testing.T) {
	// Sensor failure: some fields are null
	data := []byte(`{
		"serial_number": "ST-00000512",
		"type": "obs_st",
		"hub_sn": "HB-00013030",
		"obs": [[1588948614, null, 0.22, null, 144, 6, 1017.57, null, 50.26, 328, 0.03, 3, 0.0, 0, 0, 0, 2.410, 1]],
		"firmware_revision": 129
	}`)

	_, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obs := parsed.(*Observation)
	if obs.WindLull != nil {
		t.Errorf("expected nil wind_lull, got %v", *obs.WindLull)
	}
	if obs.WindGust != nil {
		t.Errorf("expected nil wind_gust, got %v", *obs.WindGust)
	}
	if obs.AirTemperature != nil {
		t.Errorf("expected nil air_temperature, got %v", *obs.AirTemperature)
	}
	if *obs.WindAvg != 0.22 {
		t.Errorf("wind_avg = %v", *obs.WindAvg)
	}
}

func TestParseRapidWind(t *testing.T) {
	data := []byte(`{
		"serial_number": "SK-00008453",
		"type": "rapid_wind",
		"hub_sn": "HB-00000001",
		"ob": [1493322445, 2.3, 128]
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "rapid_wind" {
		t.Fatalf("expected rapid_wind, got %s", msgType)
	}

	rw := parsed.(*RapidWind)
	if rw.WindSpeed != 2.3 {
		t.Errorf("wind_speed = %v", rw.WindSpeed)
	}
	if rw.WindDirection != 128 {
		t.Errorf("wind_direction = %v", rw.WindDirection)
	}
}

func TestParseEvtPrecip(t *testing.T) {
	data := []byte(`{
		"serial_number": "SK-00008453",
		"type": "evt_precip",
		"hub_sn": "HB-00000001",
		"evt": [1493322445]
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "evt_precip" {
		t.Fatalf("expected evt_precip, got %s", msgType)
	}

	evt := parsed.(*RainEvent)
	if evt.Time != time.Unix(1493322445, 0).UTC() {
		t.Errorf("time = %v", evt.Time)
	}
}

func TestParseEvtStrike(t *testing.T) {
	data := []byte(`{
		"serial_number": "AR-00004049",
		"type": "evt_strike",
		"hub_sn": "HB-00000001",
		"evt": [1493322445, 27, 3848]
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "evt_strike" {
		t.Fatalf("expected evt_strike, got %s", msgType)
	}

	evt := parsed.(*LightningEvent)
	if evt.Distance != 27 {
		t.Errorf("distance = %v", evt.Distance)
	}
	if evt.Energy != 3848 {
		t.Errorf("energy = %v", evt.Energy)
	}
}

func TestParseDeviceStatus(t *testing.T) {
	data := []byte(`{
		"serial_number": "AR-00004049",
		"type": "device_status",
		"hub_sn": "HB-00000001",
		"timestamp": 1510855923,
		"uptime": 2189,
		"voltage": 3.50,
		"firmware_revision": 17,
		"rssi": -17,
		"hub_rssi": -87,
		"sensor_status": 0,
		"debug": 0
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "device_status" {
		t.Fatalf("expected device_status, got %s", msgType)
	}

	ds := parsed.(*DeviceStatus)
	if ds.Voltage != 3.50 {
		t.Errorf("voltage = %v", ds.Voltage)
	}
	if ds.RSSI != -17 {
		t.Errorf("rssi = %v", ds.RSSI)
	}
	if ds.HubRSSI != -87 {
		t.Errorf("hub_rssi = %v", ds.HubRSSI)
	}
	if ds.SensorStatus != 0 {
		t.Errorf("sensor_status = %v", ds.SensorStatus)
	}
}

func TestParseHubStatus(t *testing.T) {
	data := []byte(`{
		"serial_number": "HB-00000001",
		"type": "hub_status",
		"firmware_revision": "35",
		"uptime": 1670133,
		"rssi": -62,
		"timestamp": 1495724691,
		"reset_flags": "BOR,PIN,POR",
		"seq": 48,
		"fs": [1, 0, 15675411, 524288],
		"radio_stats": [2, 1, 0, 3, 2839],
		"mqtt_stats": [1, 0]
	}`)

	msgType, parsed, err := ParsePacket(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgType != "hub_status" {
		t.Fatalf("expected hub_status, got %s", msgType)
	}

	hs := parsed.(*HubStatus)
	if hs.FirmwareRevision != "35" {
		t.Errorf("firmware_revision = %q", hs.FirmwareRevision)
	}
	if hs.ResetFlags != "BOR,PIN,POR" {
		t.Errorf("reset_flags = %q", hs.ResetFlags)
	}
	if hs.RadioStatus != 3 {
		t.Errorf("radio_status = %v", hs.RadioStatus)
	}
	if hs.RadioNetworkID != 2839 {
		t.Errorf("radio_network_id = %v", hs.RadioNetworkID)
	}
}

func TestParseUnknownType(t *testing.T) {
	data := []byte(`{"type": "obs_air", "serial_number": "AR-00004049", "hub_sn": "HB-00000001"}`)
	_, _, err := ParsePacket(data)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestParseMalformedJSON(t *testing.T) {
	_, _, err := ParsePacket([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for malformed json")
	}
}
