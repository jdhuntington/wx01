import { useCallback, useEffect, useMemo } from 'react';
import { getCurrent, getTemperature, getHumidity, getWind, getPressure, getRain, getSolar, getUV, getLightning } from '../api';
import type { CurrentData } from '../api';
import type { UnitSystem } from '../units';
import { tempConvert, tempUnit, windConvert, windUnit, rainConvert, rainUnit, pressureConvert, pressureUnit } from '../units';
import { useFetch } from '../hooks';
import Sparkline from './Sparkline';

function windDirLabel(deg: number | null): string {
  if (deg === null) return '--';
  const dirs = ['N','NNE','NE','ENE','E','ESE','SE','SSE','S','SSW','SW','WSW','W','WNW','NW','NNW'];
  return dirs[Math.round(deg / 22.5) % 16];
}

function fmt(v: number | null | undefined, decimals = 1, suffix = ''): string {
  if (v === null || v === undefined) return '--';
  return v.toFixed(decimals) + suffix;
}

function extractSeries(data: any[] | null, key: string, transform?: (v: number) => number): number[] {
  if (!data) return [];
  return data
    .map((d: any) => d[key])
    .filter((v: any) => v !== null && v !== undefined)
    .map((v: number) => transform ? transform(v) : v);
}

interface Props {
  units: UnitSystem;
  range_: string;
  version: number;
  onUpdate?: (time: string) => void;
}

export default function CurrentConditions({ units, range_, version, onUpdate }: Props) {
  const fetcher = useCallback(() => getCurrent(), []);
  const { data, error } = useFetch<CurrentData>(fetcher, version);

  const tempFetcher = useCallback(() => getTemperature(range_), [range_]);
  const humidityFetcher = useCallback(() => getHumidity(range_), [range_]);
  const windFetcher = useCallback(() => getWind(range_), [range_]);
  const pressureFetcher = useCallback(() => getPressure(range_), [range_]);
  const rainFetcher = useCallback(() => getRain(range_), [range_]);
  const solarFetcher = useCallback(() => getSolar(range_), [range_]);
  const uvFetcher = useCallback(() => getUV(range_), [range_]);
  const lightningFetcher = useCallback(() => getLightning(range_), [range_]);

  const { data: tempData } = useFetch(tempFetcher, version);
  const { data: humidityData } = useFetch(humidityFetcher, version);
  const { data: windData } = useFetch(windFetcher, version);
  const { data: pressureData } = useFetch(pressureFetcher, version);
  const { data: rainData } = useFetch(rainFetcher, version);
  const { data: solarData } = useFetch(solarFetcher, version);
  const { data: uvData } = useFetch(uvFetcher, version);
  const { data: lightningData } = useFetch(lightningFetcher, version);

  const tempTransform = useMemo(() => (v: number) => tempConvert(v, units)!, [units]);
  const windTransform = useMemo(() => (v: number) => windConvert(v, units)!, [units]);
  const rainTransform = useMemo(() => (v: number) => rainConvert(v, units)!, [units]);
  const pressureTransform = useMemo(() => (v: number) => pressureConvert(v, units)!, [units]);

  const tempSeries = useMemo(() => extractSeries(tempData, 'temp_avg_c', tempTransform), [tempData, tempTransform]);
  const humiditySeries = useMemo(() => extractSeries(humidityData, 'humidity_avg_pct'), [humidityData]);
  const windSeries = useMemo(() => extractSeries(windData, 'wind_avg_ms', windTransform), [windData, windTransform]);
  const pressureSeries = useMemo(() => extractSeries(pressureData, 'pressure_avg_mb', pressureTransform), [pressureData, pressureTransform]);
  const rainSeries = useMemo(() => extractSeries(rainData, 'rain_mm', rainTransform), [rainData, rainTransform]);
  const solarSeries = useMemo(() => extractSeries(solarData, 'solar_avg_wm2'), [solarData]);
  const uvSeries = useMemo(() => extractSeries(uvData, 'uv_avg'), [uvData]);
  const lightningSeries = useMemo(() => extractSeries(lightningData, 'strike_count'), [lightningData]);
  const lightningTotal = useMemo(
    () => (lightningData as any[] | null)?.reduce((sum, b) => sum + (b.strike_count ?? 0), 0) ?? 0,
    [lightningData],
  );

  useEffect(() => {
    if (data?.observation?.time && onUpdate) {
      onUpdate(data.observation.time);
    }
  }, [data, onUpdate]);

  if (error) return <div className="card error">Failed to load current conditions</div>;
  if (!data) return <div className="current-grid">{Array.from({length: 7}, (_, i) => <div key={i} className="card metric skeleton" />)}</div>;

  const o = data.observation;

  return (
    <div className="current-grid">
      <div className="card metric hero-metric">
        <div className="label">Temperature</div>
        <div className="value hero">{fmt(tempConvert(o.air_temperature, units), 1, tempUnit(units))}</div>
        <Sparkline data={tempSeries} color="#e74c3c" />
      </div>
      <div className="card metric">
        <div className="label">Humidity</div>
        <div className="value">{fmt(o.relative_humidity, 0, '%')}</div>
        <Sparkline data={humiditySeries} color="#3498db" />
      </div>
      <div className="card metric">
        <div className="label">Wind</div>
        <div className="value">{fmt(windConvert(o.wind_avg, units), 1, ' ' + windUnit(units))}</div>
        <div className="detail">
          Gust {fmt(windConvert(o.wind_gust, units), 1)} · {windDirLabel(o.wind_direction)} {o.wind_direction}°
        </div>
        <Sparkline data={windSeries} color="#2ecc71" />
      </div>
      <div className="card metric">
        <div className="label">Pressure</div>
        <div className="value">{fmt(pressureConvert(o.station_pressure, units), units === 'imperial' ? 2 : 1, ' ' + pressureUnit(units))}</div>
        <Sparkline data={pressureSeries} color="#9b59b6" />
      </div>
      <div className="card metric">
        <div className="label">Rain</div>
        <div className="value">{fmt(rainConvert(data.rain_last_hour, units), units === 'imperial' ? 2 : 1, ' ' + rainUnit(units))}<span className="sub"> /hr</span></div>
        <div className="detail">{fmt(rainConvert(data.rain_today, units), units === 'imperial' ? 2 : 1, ' ' + rainUnit(units))} today</div>
        <Sparkline data={rainSeries} color="#3498db" />
      </div>
      <div className="card metric">
        <div className="label">Solar</div>
        <div className="value">{fmt(o.solar_radiation, 0, ' W/m²')}</div>
        <Sparkline data={solarSeries} color="#f39c12" />
      </div>
      <div className="card metric">
        <div className="label">UV Index</div>
        <div className="value">{fmt(o.uv, 1)}</div>
        <Sparkline data={uvSeries} color="#e67e22" />
      </div>
      {(lightningTotal > 0 || data.lightning_last_hour > 0) && (
        <div className="card metric">
          <div className="label">Lightning</div>
          <div className="value">
            {data.lightning_last_hour}
            <span className="sub"> /hr</span>
          </div>
          <div className="detail">
            {data.lightning_closest_km !== null
              ? `Closest ${data.lightning_closest_km} km · ${lightningTotal} in range`
              : `${lightningTotal} in range`}
          </div>
          <Sparkline data={lightningSeries} color="#f1c40f" />
        </div>
      )}
    </div>
  );
}
