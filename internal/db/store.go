package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jedediah/wx01/internal/ingest"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) notify(ctx context.Context, msgType string) {
	s.pool.Exec(ctx, "SELECT pg_notify('wx01_data', $1)", msgType)
}

func (s *Store) InsertObservation(ctx context.Context, o *ingest.Observation) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO observations (
			time, serial_number, hub_sn,
			wind_lull, wind_avg, wind_gust, wind_direction, wind_sample_interval,
			station_pressure, air_temperature, relative_humidity,
			illuminance, uv, solar_radiation,
			rain_accumulated, precipitation_type,
			lightning_avg_distance, lightning_strike_count,
			battery, report_interval, firmware_revision
		) VALUES (
			$1, $2, $3,
			$4, $5, $6, $7, $8,
			$9, $10, $11,
			$12, $13, $14,
			$15, $16,
			$17, $18,
			$19, $20, $21
		)`,
		o.Time, o.SerialNumber, o.HubSN,
		o.WindLull, o.WindAvg, o.WindGust, o.WindDirection, o.WindSampleInterval,
		o.StationPressure, o.AirTemperature, o.RelativeHumidity,
		o.Illuminance, o.UV, o.SolarRadiation,
		o.RainAccumulated, o.PrecipitationType,
		o.LightningAvgDist, o.LightningCount,
		o.Battery, o.ReportInterval, o.FirmwareRevision,
	)
	if err == nil {
		s.notify(ctx, "obs_st")
	}
	return err
}

func (s *Store) InsertRapidWind(ctx context.Context, rw *ingest.RapidWind) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO rapid_wind (time, serial_number, hub_sn, wind_speed, wind_direction)
		VALUES ($1, $2, $3, $4, $5)`,
		rw.Time, rw.SerialNumber, rw.HubSN, rw.WindSpeed, rw.WindDirection,
	)
	if err == nil {
		s.notify(ctx, "rapid_wind")
	}
	return err
}

func (s *Store) InsertRainEvent(ctx context.Context, evt *ingest.RainEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO rain_events (time, serial_number, hub_sn)
		VALUES ($1, $2, $3)`,
		evt.Time, evt.SerialNumber, evt.HubSN,
	)
	if err == nil {
		s.notify(ctx, "evt_precip")
	}
	return err
}

func (s *Store) InsertLightningEvent(ctx context.Context, evt *ingest.LightningEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO lightning_events (time, serial_number, hub_sn, distance, energy)
		VALUES ($1, $2, $3, $4, $5)`,
		evt.Time, evt.SerialNumber, evt.HubSN, evt.Distance, evt.Energy,
	)
	if err == nil {
		s.notify(ctx, "evt_strike")
	}
	return err
}

func (s *Store) InsertDeviceStatus(ctx context.Context, ds *ingest.DeviceStatus) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO device_status (
			time, serial_number, hub_sn,
			uptime, voltage, firmware_revision,
			rssi, hub_rssi, sensor_status, debug
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		ds.Time, ds.SerialNumber, ds.HubSN,
		ds.Uptime, ds.Voltage, ds.FirmwareRevision,
		ds.RSSI, ds.HubRSSI, ds.SensorStatus, ds.Debug,
	)
	if err == nil {
		s.notify(ctx, "device_status")
	}
	return err
}

func (s *Store) InsertHubStatus(ctx context.Context, hs *ingest.HubStatus) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO hub_status (
			time, serial_number, firmware_revision,
			uptime, rssi, reset_flags, seq,
			radio_status, radio_reboot_count, radio_i2c_errors, radio_network_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		hs.Time, hs.SerialNumber, hs.FirmwareRevision,
		hs.Uptime, hs.RSSI, hs.ResetFlags, hs.Seq,
		hs.RadioStatus, hs.RadioRebootCount, hs.RadioI2CErrors, hs.RadioNetworkID,
	)
	if err == nil {
		s.notify(ctx, "hub_status")
	}
	return err
}
