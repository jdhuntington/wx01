-- Derived views. All read from raw tables, nothing materialized yet.
-- Can promote to continuous aggregates later if query performance demands it.

-- Current conditions: latest observation
CREATE VIEW current_conditions AS
SELECT *
FROM observations
ORDER BY time DESC
LIMIT 1;

-- Hourly rain totals
CREATE VIEW hourly_rain AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    sum(rain_accumulated)       AS rain_mm
FROM observations
GROUP BY bucket
ORDER BY bucket DESC;

-- Daily rain totals
CREATE VIEW daily_rain AS
SELECT
    time_bucket('1 day', time) AS bucket,
    sum(rain_accumulated)      AS rain_mm
FROM observations
GROUP BY bucket
ORDER BY bucket DESC;

-- Hourly wind summary
CREATE VIEW hourly_wind AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    avg(wind_avg)               AS wind_avg_ms,
    max(wind_gust)              AS wind_gust_max_ms,
    min(wind_lull)              AS wind_lull_min_ms
FROM observations
GROUP BY bucket
ORDER BY bucket DESC;

-- Hourly temperature and humidity
CREATE VIEW hourly_temp_humidity AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    avg(air_temperature)        AS temp_avg_c,
    min(air_temperature)        AS temp_min_c,
    max(air_temperature)        AS temp_max_c,
    avg(relative_humidity)      AS humidity_avg_pct
FROM observations
GROUP BY bucket
ORDER BY bucket DESC;

-- Pressure trend: hourly average pressure
CREATE VIEW hourly_pressure AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    avg(station_pressure)       AS pressure_avg_mb
FROM observations
GROUP BY bucket
ORDER BY bucket DESC;

-- Lightning activity: hourly strike counts and nearest distance
CREATE VIEW hourly_lightning AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    count(*)                    AS strike_count,
    min(distance)               AS nearest_km
FROM lightning_events
GROUP BY bucket
ORDER BY bucket DESC;

-- Rapid wind summary: 5-minute averages for when you want higher resolution than obs
CREATE VIEW rapid_wind_5min AS
SELECT
    time_bucket('5 minutes', time) AS bucket,
    avg(wind_speed)                AS wind_avg_ms,
    max(wind_speed)                AS wind_max_ms,
    min(wind_speed)                AS wind_min_ms
FROM rapid_wind
GROUP BY bucket
ORDER BY bucket DESC;

-- Latest device health
CREATE VIEW current_device_status AS
SELECT *
FROM device_status
ORDER BY time DESC
LIMIT 1;

-- Latest hub health
CREATE VIEW current_hub_status AS
SELECT *
FROM hub_status
ORDER BY time DESC
LIMIT 1;
