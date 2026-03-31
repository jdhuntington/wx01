const BASE = '/api';

export async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
  return res.json();
}

export interface CurrentData {
  observation: {
    time: string;
    air_temperature: number | null;
    relative_humidity: number | null;
    station_pressure: number | null;
    wind_avg: number | null;
    wind_gust: number | null;
    wind_lull: number | null;
    wind_direction: number | null;
    rain_accumulated: number | null;
    uv: number | null;
    solar_radiation: number | null;
    illuminance: number | null;
    lightning_strike_count: number | null;
    lightning_avg_distance: number | null;
    battery: number | null;
  };
  rain_last_hour: number;
  rain_today: number;
}

export interface TempBucket {
  bucket: string;
  temp_avg_c: number | null;
  temp_min_c: number | null;
  temp_max_c: number | null;
  humidity_avg_pct: number | null;
}

export interface WindBucket {
  bucket: string;
  wind_avg_ms: number | null;
  wind_gust_max_ms: number | null;
  wind_lull_min_ms: number | null;
}

export interface RainBucket {
  bucket: string;
  rain_mm: number | null;
}

export interface PressureBucket {
  bucket: string;
  pressure_avg_mb: number | null;
}

export interface SolarBucket {
  bucket: string;
  solar_avg_wm2: number | null;
  solar_max_wm2: number | null;
}

export interface HumidityBucket {
  bucket: string;
  humidity_avg_pct: number | null;
  humidity_min_pct: number | null;
  humidity_max_pct: number | null;
}

export const getCurrent = () => fetchJSON<CurrentData>('/current');
export const getTemperature = (range_: string) => fetchJSON<TempBucket[]>(`/temperature?range=${range_}`);
export const getWind = (range_: string) => fetchJSON<WindBucket[]>(`/wind?range=${range_}`);
export const getRain = (range_: string) => fetchJSON<RainBucket[]>(`/rain?range=${range_}`);
export const getPressure = (range_: string) => fetchJSON<PressureBucket[]>(`/pressure?range=${range_}`);
export const getSolar = (range_: string) => fetchJSON<SolarBucket[]>(`/solar?range=${range_}`);
export const getHumidity = (range_: string) => fetchJSON<HumidityBucket[]>(`/humidity?range=${range_}`);

export interface UVBucket {
  bucket: string;
  uv_avg: number | null;
  uv_max: number | null;
}

export const getUV = (range_: string) => fetchJSON<UVBucket[]>(`/uv?range=${range_}`);
