package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CurrentConditions struct {
	Time             time.Time `json:"time"`
	AirTemperature   *float64  `json:"air_temperature"`
	RelativeHumidity *float64  `json:"relative_humidity"`
	StationPressure  *float64  `json:"station_pressure"`
	WindAvg          *float64  `json:"wind_avg"`
	WindGust         *float64  `json:"wind_gust"`
	WindLull         *float64  `json:"wind_lull"`
	WindDirection    *int16    `json:"wind_direction"`
	RainAccumulated  *float64  `json:"rain_accumulated"`
	UV               *float64  `json:"uv"`
	SolarRadiation   *int32    `json:"solar_radiation"`
	Illuminance      *int32    `json:"illuminance"`
	LightningCount   *int16    `json:"lightning_strike_count"`
	LightningDist    *int16    `json:"lightning_avg_distance"`
	Battery          *float64  `json:"battery"`
}

type TimeBucket struct {
	Bucket time.Time `json:"bucket"`
}

type TempHumidityBucket struct {
	TimeBucket
	TempAvg    *float64 `json:"temp_avg_c"`
	TempMin    *float64 `json:"temp_min_c"`
	TempMax    *float64 `json:"temp_max_c"`
	HumidityAvg *float64 `json:"humidity_avg_pct"`
}

type WindBucket struct {
	TimeBucket
	WindAvg     *float64 `json:"wind_avg_ms"`
	WindGustMax *float64 `json:"wind_gust_max_ms"`
	WindLullMin *float64 `json:"wind_lull_min_ms"`
}

type RainBucket struct {
	TimeBucket
	RainMM *float64 `json:"rain_mm"`
}

type PressureBucket struct {
	TimeBucket
	PressureAvg *float64 `json:"pressure_avg_mb"`
}

type SolarBucket struct {
	TimeBucket
	SolarAvg *float64 `json:"solar_avg_wm2"`
	SolarMax *float64 `json:"solar_max_wm2"`
}

type HumidityBucket struct {
	TimeBucket
	HumidityAvg *float64 `json:"humidity_avg_pct"`
	HumidityMin *float64 `json:"humidity_min_pct"`
	HumidityMax *float64 `json:"humidity_max_pct"`
}

type UVBucket struct {
	TimeBucket
	UVAvg *float64 `json:"uv_avg"`
	UVMax *float64 `json:"uv_max"`
}

type LightningBucket struct {
	TimeBucket
	StrikeCount int      `json:"strike_count"`
	DistanceMin *float64 `json:"distance_min_km"`
	DistanceMax *float64 `json:"distance_max_km"`
	EnergyMax   *int64   `json:"energy_max"`
}

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) CurrentConditions(ctx context.Context) (*CurrentConditions, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT time, air_temperature, relative_humidity, station_pressure,
			wind_avg, wind_gust, wind_lull, wind_direction,
			rain_accumulated, uv, solar_radiation, illuminance,
			lightning_strike_count, lightning_avg_distance, battery
		FROM observations ORDER BY time DESC LIMIT 1
	`)
	var c CurrentConditions
	err := row.Scan(
		&c.Time, &c.AirTemperature, &c.RelativeHumidity, &c.StationPressure,
		&c.WindAvg, &c.WindGust, &c.WindLull, &c.WindDirection,
		&c.RainAccumulated, &c.UV, &c.SolarRadiation, &c.Illuminance,
		&c.LightningCount, &c.LightningDist, &c.Battery,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (q *Queries) TempHumidity(ctx context.Context, since time.Time, interval string) ([]TempHumidityBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(air_temperature), min(air_temperature), max(air_temperature),
			avg(relative_humidity)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TempHumidityBucket
	for rows.Next() {
		var b TempHumidityBucket
		if err := rows.Scan(&b.Bucket, &b.TempAvg, &b.TempMin, &b.TempMax, &b.HumidityAvg); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Wind(ctx context.Context, since time.Time, interval string) ([]WindBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(wind_avg), max(wind_gust), min(wind_lull)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []WindBucket
	for rows.Next() {
		var b WindBucket
		if err := rows.Scan(&b.Bucket, &b.WindAvg, &b.WindGustMax, &b.WindLullMin); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Rain(ctx context.Context, since time.Time, interval string) ([]RainBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			sum(rain_accumulated)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RainBucket
	for rows.Next() {
		var b RainBucket
		if err := rows.Scan(&b.Bucket, &b.RainMM); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Pressure(ctx context.Context, since time.Time, interval string) ([]PressureBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(station_pressure)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PressureBucket
	for rows.Next() {
		var b PressureBucket
		if err := rows.Scan(&b.Bucket, &b.PressureAvg); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Solar(ctx context.Context, since time.Time, interval string) ([]SolarBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(solar_radiation), max(solar_radiation)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SolarBucket
	for rows.Next() {
		var b SolarBucket
		if err := rows.Scan(&b.Bucket, &b.SolarAvg, &b.SolarMax); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Humidity(ctx context.Context, since time.Time, interval string) ([]HumidityBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(relative_humidity), min(relative_humidity), max(relative_humidity)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HumidityBucket
	for rows.Next() {
		var b HumidityBucket
		if err := rows.Scan(&b.Bucket, &b.HumidityAvg, &b.HumidityMin, &b.HumidityMax); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) UV(ctx context.Context, since time.Time, interval string) ([]UVBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			avg(uv), max(uv)
		FROM observations WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UVBucket
	for rows.Next() {
		var b UVBucket
		if err := rows.Scan(&b.Bucket, &b.UVAvg, &b.UVMax); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

func (q *Queries) Lightning(ctx context.Context, since time.Time, interval string) ([]LightningBucket, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time_bucket($1::interval, time) AS bucket,
			count(*)::int, min(distance)::double precision, max(distance)::double precision, max(energy)::bigint
		FROM lightning_events WHERE time >= $2
		GROUP BY bucket ORDER BY bucket
	`, interval, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []LightningBucket
	for rows.Next() {
		var b LightningBucket
		if err := rows.Scan(&b.Bucket, &b.StrikeCount, &b.DistanceMin, &b.DistanceMax, &b.EnergyMax); err != nil {
			return nil, err
		}
		result = append(result, b)
	}
	return result, rows.Err()
}

type LightningStrike struct {
	Time     time.Time `json:"time"`
	Distance int16     `json:"distance_km"`
	Energy   int32     `json:"energy"`
}

func (q *Queries) LightningStrikes(ctx context.Context, since time.Time) ([]LightningStrike, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT time, distance, energy
		FROM lightning_events WHERE time >= $1
		ORDER BY time
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []LightningStrike
	for rows.Next() {
		var s LightningStrike
		if err := rows.Scan(&s.Time, &s.Distance, &s.Energy); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func (q *Queries) LightningLastHour(ctx context.Context) (int, *float64, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT count(*)::int, min(distance)::double precision
		FROM lightning_events
		WHERE time >= now() - interval '1 hour'
	`)
	var count int
	var minDist *float64
	err := row.Scan(&count, &minDist)
	return count, minDist, err
}

func (q *Queries) RainToday(ctx context.Context) (float64, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT COALESCE(sum(rain_accumulated), 0)
		FROM observations
		WHERE time >= date_trunc('day', now())
	`)
	var total float64
	err := row.Scan(&total)
	return total, err
}

func (q *Queries) RainLastHour(ctx context.Context) (float64, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT COALESCE(sum(rain_accumulated), 0)
		FROM observations
		WHERE time >= now() - interval '1 hour'
	`)
	var total float64
	err := row.Scan(&total)
	return total, err
}
