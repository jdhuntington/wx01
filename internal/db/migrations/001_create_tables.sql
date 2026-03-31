-- wx01 schema v1
-- One table per Tempest UDP message type. Store raw, derive later.

CREATE EXTENSION IF NOT EXISTS timescaledb;

-- obs_st: once per minute, all sensor readings
CREATE TABLE observations (
    time               timestamptz NOT NULL,
    serial_number      text        NOT NULL,
    hub_sn             text        NOT NULL,
    wind_lull          double precision,  -- m/s, min 3s sample
    wind_avg           double precision,  -- m/s, avg over interval
    wind_gust          double precision,  -- m/s, max 3s sample
    wind_direction     smallint,          -- degrees
    wind_sample_interval smallint,        -- seconds
    station_pressure   double precision,  -- mb
    air_temperature    double precision,  -- celsius
    relative_humidity  double precision,  -- percent
    illuminance        integer,           -- lux
    uv                 double precision,  -- index
    solar_radiation    integer,           -- W/m^2
    rain_accumulated   double precision,  -- mm over previous minute
    precipitation_type smallint,          -- 0=none, 1=rain, 2=hail, 3=rain+hail
    lightning_avg_distance smallint,      -- km
    lightning_strike_count smallint,      -- count
    battery            double precision,  -- volts
    report_interval    smallint,          -- minutes
    firmware_revision  integer
);

SELECT create_hypertable('observations', 'time');
CREATE INDEX idx_observations_time ON observations (time DESC);

-- rapid_wind: every 3 seconds
CREATE TABLE rapid_wind (
    time           timestamptz      NOT NULL,
    serial_number  text             NOT NULL,
    hub_sn         text             NOT NULL,
    wind_speed     double precision NOT NULL, -- m/s
    wind_direction smallint         NOT NULL  -- degrees
);

SELECT create_hypertable('rapid_wind', 'time');
CREATE INDEX idx_rapid_wind_time ON rapid_wind (time DESC);

-- evt_precip: rain start event (timestamp only)
CREATE TABLE rain_events (
    time          timestamptz NOT NULL,
    serial_number text        NOT NULL,
    hub_sn        text        NOT NULL
);

SELECT create_hypertable('rain_events', 'time');

-- evt_strike: lightning strike event
CREATE TABLE lightning_events (
    time          timestamptz NOT NULL,
    serial_number text        NOT NULL,
    hub_sn        text        NOT NULL,
    distance      smallint    NOT NULL, -- km
    energy        integer     NOT NULL  -- unitless
);

SELECT create_hypertable('lightning_events', 'time');

-- device_status: ~once per minute
CREATE TABLE device_status (
    time              timestamptz NOT NULL,
    serial_number     text        NOT NULL,
    hub_sn            text        NOT NULL,
    uptime            integer,            -- seconds
    voltage           double precision,   -- volts
    firmware_revision integer,
    rssi              smallint,           -- dBm
    hub_rssi          smallint,           -- dBm
    sensor_status     integer,            -- bitmask
    debug             smallint
);

SELECT create_hypertable('device_status', 'time');

-- hub_status: ~once per minute
CREATE TABLE hub_status (
    time              timestamptz NOT NULL,
    serial_number     text        NOT NULL,
    firmware_revision text,               -- string, not int
    uptime            integer,            -- seconds
    rssi              smallint,           -- dBm, wifi
    reset_flags       text,               -- comma-separated codes
    seq               integer,            -- sequence number
    radio_status      smallint,           -- 0=off, 1=on, 3=active, 7=BLE
    radio_reboot_count integer,
    radio_i2c_errors  integer,
    radio_network_id  integer
);

SELECT create_hypertable('hub_status', 'time');
