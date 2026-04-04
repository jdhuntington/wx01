import { useState, useCallback, useMemo, useEffect } from 'react';
import CurrentConditions from './components/CurrentConditions';
import TimeSeriesChart from './components/TimeSeriesChart';
import { getTemperature, getWind, getRain, getPressure, getSolar, getHumidity, getUV } from './api';
import { getStoredUnits, setStoredUnits, tempConvert, tempUnit, windConvert, windUnit, rainConvert, rainUnit, pressureConvert, pressureUnit } from './units';
import type { UnitSystem } from './units';
import { useServerEvents } from './hooks';

const RANGES = [
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '168h' },
  { label: '30d', value: '720h' },
];

function timeAgo(iso: string): string {
  const s = Math.round((Date.now() - new Date(iso).getTime()) / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  return `${Math.floor(s / 3600)}h ago`;
}

function useCompact(query = '(max-width: 400px)') {
  const [match, setMatch] = useState(() => window.matchMedia(query).matches);
  useEffect(() => {
    const mql = window.matchMedia(query);
    const handler = (e: MediaQueryListEvent) => setMatch(e.matches);
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [query]);
  return match;
}

export default function App() {
  const compact = useCompact();
  const version = useServerEvents();
  const [range_, setRange] = useState('24h');
  const [units, setUnits] = useState<UnitSystem>(getStoredUnits);
  const [lastUpdate, setLastUpdate] = useState<string | null>(null);
  const [, setTick] = useState(0);

  // Re-render every 5s so "time ago" stays fresh
  useState(() => {
    const id = setInterval(() => setTick(t => t + 1), 5000);
    return () => clearInterval(id);
  });

  const toggleUnits = useCallback(() => {
    setUnits(prev => {
      const next = prev === 'metric' ? 'imperial' : 'metric';
      setStoredUnits(next);
      return next;
    });
  }, []);

  const handleUpdate = useCallback((time: string) => {
    setLastUpdate(time);
  }, []);

  const tempTransform = useMemo(() => (v: number) => tempConvert(v, units)!, [units]);
  const windTransform = useMemo(() => (v: number) => windConvert(v, units)!, [units]);
  const rainTransform = useMemo(() => (v: number) => rainConvert(v, units)!, [units]);
  const pressureTransform = useMemo(() => (v: number) => pressureConvert(v, units)!, [units]);

  const tempFormatter = useMemo(() => (v: number) => `${v.toFixed(1)}${tempUnit(units)}`, [units]);
  const windFormatter = useMemo(() => (v: number) => `${v.toFixed(1)} ${windUnit(units)}`, [units]);
  const rainFormatter = useMemo(() => {
    const dec = units === 'imperial' ? 2 : 1;
    return (v: number) => `${v.toFixed(dec)} ${rainUnit(units)}`;
  }, [units]);
  const pressureFormatter = useMemo(() => {
    const dec = units === 'imperial' ? 2 : 1;
    return (v: number) => `${v.toFixed(dec)} ${pressureUnit(units)}`;
  }, [units]);

  return (
    <div className="app">
      <header>
        <div className="header-left">
          <h1>wx01</h1>
          {lastUpdate && (
            <span className="last-update">{timeAgo(lastUpdate)}</span>
          )}
        </div>
        <div className="header-controls">
          <div className="range-picker">
            {RANGES.map(r => (
              <button
                key={r.value}
                className={range_ === r.value ? 'active' : ''}
                onClick={() => setRange(r.value)}
              >
                {r.label}
              </button>
            ))}
          </div>
          <button className="unit-toggle" onClick={toggleUnits}>
            {units === 'metric' ? '°C' : '°F'}
          </button>
        </div>
      </header>

      <CurrentConditions units={units} range_={range_} version={version} onUpdate={handleUpdate} />

      <div className="charts">
        <TimeSeriesChart
          title="Temperature"
          fetcher={getTemperature}
          range_={range_}
          series={[{ name: 'Temp', key: 'temp_avg_c', color: '#e74c3c' }]}
          yFormat={tempFormatter}
          transform={tempTransform}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="Humidity"
          fetcher={getHumidity}
          range_={range_}
          series={[{ name: 'Humidity', key: 'humidity_avg_pct', color: '#3498db' }]}
          yFormat={(v) => `${v.toFixed(0)}%`}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="Wind"
          fetcher={getWind}
          range_={range_}
          series={[
            { name: 'Avg', key: 'wind_avg_ms', color: '#2ecc71' },
            { name: 'Gust', key: 'wind_gust_max_ms', color: '#e74c3c' },
          ]}
          yFormat={windFormatter}
          transform={windTransform}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="Pressure"
          fetcher={getPressure}
          range_={range_}
          series={[{ name: 'Pressure', key: 'pressure_avg_mb', color: '#9b59b6' }]}
          yFormat={pressureFormatter}
          transform={pressureTransform}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="Rain"
          fetcher={getRain}
          range_={range_}
          series={[{ name: 'Rain', key: 'rain_mm', color: '#3498db', type: 'histogram' }]}
          yFormat={rainFormatter}
          transform={rainTransform}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="Solar Radiation"
          fetcher={getSolar}
          range_={range_}
          series={[{ name: 'Solar', key: 'solar_avg_wm2', color: '#f39c12' }]}
          yFormat={(v) => `${v.toFixed(0)} W/m²`}
          compact={compact}
          version={version}
        />

        <TimeSeriesChart
          title="UV Index"
          fetcher={getUV}
          range_={range_}
          series={[{ name: 'UV', key: 'uv_avg', color: '#e67e22' }]}
          yFormat={(v) => `${v.toFixed(1)}`}
          compact={compact}
          version={version}
        />
      </div>
    </div>
  );
}
