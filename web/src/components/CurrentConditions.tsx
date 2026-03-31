import { useCallback, useEffect } from 'react';
import { getCurrent } from '../api';
import type { CurrentData } from '../api';
import type { UnitSystem } from '../units';
import { tempConvert, tempUnit, windConvert, windUnit, rainConvert, rainUnit, pressureConvert, pressureUnit } from '../units';
import { usePolling } from '../hooks';

function windDirLabel(deg: number | null): string {
  if (deg === null) return '--';
  const dirs = ['N','NNE','NE','ENE','E','ESE','SE','SSE','S','SSW','SW','WSW','W','WNW','NW','NNW'];
  return dirs[Math.round(deg / 22.5) % 16];
}

function fmt(v: number | null | undefined, decimals = 1, suffix = ''): string {
  if (v === null || v === undefined) return '--';
  return v.toFixed(decimals) + suffix;
}

interface Props {
  units: UnitSystem;
  onUpdate?: (time: string) => void;
}

export default function CurrentConditions({ units, onUpdate }: Props) {
  const fetcher = useCallback(() => getCurrent(), []);
  const { data, error } = usePolling<CurrentData>(fetcher, 10_000);

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
      </div>
      <div className="card metric">
        <div className="label">Humidity</div>
        <div className="value">{fmt(o.relative_humidity, 0, '%')}</div>
      </div>
      <div className="card metric">
        <div className="label">Wind</div>
        <div className="value">{fmt(windConvert(o.wind_avg, units), 1, ' ' + windUnit(units))}</div>
        <div className="detail">
          Gust {fmt(windConvert(o.wind_gust, units), 1)} · {windDirLabel(o.wind_direction)} {o.wind_direction}°
        </div>
      </div>
      <div className="card metric">
        <div className="label">Pressure</div>
        <div className="value">{fmt(pressureConvert(o.station_pressure, units), units === 'imperial' ? 2 : 1, ' ' + pressureUnit(units))}</div>
      </div>
      <div className="card metric">
        <div className="label">Rain</div>
        <div className="value">{fmt(rainConvert(data.rain_last_hour, units), units === 'imperial' ? 2 : 1, ' ' + rainUnit(units))}<span className="sub"> /hr</span></div>
        <div className="detail">{fmt(rainConvert(data.rain_today, units), units === 'imperial' ? 2 : 1, ' ' + rainUnit(units))} today</div>
      </div>
      <div className="card metric">
        <div className="label">Solar</div>
        <div className="value">{fmt(o.solar_radiation, 0, ' W/m²')}</div>
      </div>
      <div className="card metric">
        <div className="label">UV Index</div>
        <div className="value">{fmt(o.uv, 1)}</div>
      </div>
    </div>
  );
}
