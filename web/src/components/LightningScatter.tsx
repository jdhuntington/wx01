import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { getLightningStrikes } from '../api';
import type { LightningStrike } from '../api';
import { useFetch } from '../hooks';
import { distanceConvert, distanceUnit } from '../units';
import type { UnitSystem } from '../units';

interface Props {
  range_: string;
  units: UnitSystem;
  version: number;
  compact?: boolean;
}

interface HoverState {
  strike: LightningStrike;
  x: number;
  y: number;
}

const PAD_LEFT = 44;
const PAD_RIGHT = 12;
const PAD_TOP = 10;
const PAD_BOTTOM = 28;

function rangeHours(range_: string): number {
  return parseInt(range_) || 24;
}

function fmtTimeAxis(date: Date, hours: number): string {
  if (hours <= 24) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return date.toLocaleDateString([], { month: 'numeric', day: 'numeric' });
}

function fmtTooltipTime(date: Date): string {
  return date.toLocaleString([], {
    month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  });
}

// Map energy (~16M-17M range observed) to a 0..1 intensity
function energyIntensity(energy: number, minE: number, maxE: number): number {
  if (maxE === minE) return 0.5;
  return Math.max(0, Math.min(1, (energy - minE) / (maxE - minE)));
}

function intensityColor(t: number): string {
  // yellow → orange → red as energy climbs
  if (t < 0.5) {
    const k = t / 0.5;
    const r = 241 + (230 - 241) * k;
    const g = 196 + (126 - 196) * k;
    const b = 15 + (34 - 15) * k;
    return `rgb(${r|0}, ${g|0}, ${b|0})`;
  }
  const k = (t - 0.5) / 0.5;
  const r = 230 + (231 - 230) * k;
  const g = 126 + (76 - 126) * k;
  const b = 34 + (60 - 34) * k;
  return `rgb(${r|0}, ${g|0}, ${b|0})`;
}

export default function LightningScatter({ range_, units, version, compact }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(600);
  const [hover, setHover] = useState<HoverState | null>(null);

  const fetcher = useCallback(() => getLightningStrikes(range_), [range_]);
  const { data, error } = useFetch<LightningStrike[]>(fetcher, version);

  useEffect(() => {
    if (!containerRef.current) return;
    const ro = new ResizeObserver(entries => {
      for (const e of entries) setWidth(e.contentRect.width);
    });
    ro.observe(containerRef.current);
    return () => ro.disconnect();
  }, []);

  const height = compact ? 140 : 240;
  const hours = rangeHours(range_);

  const { points, yMaxKm, yMaxDisplay, energyMin, energyMax, xStart, xEnd } = useMemo(() => {
    const xEnd = Date.now();
    const xStart = xEnd - hours * 3600 * 1000;

    const strikes = data ?? [];
    let yMaxKm = 10;
    let energyMin = Infinity;
    let energyMax = -Infinity;
    for (const s of strikes) {
      if (s.distance_km > yMaxKm) yMaxKm = s.distance_km;
      if (s.energy < energyMin) energyMin = s.energy;
      if (s.energy > energyMax) energyMax = s.energy;
    }
    if (!isFinite(energyMin)) { energyMin = 0; energyMax = 1; }
    yMaxKm = Math.ceil(yMaxKm * 1.1);

    const yMaxDisplay = distanceConvert(yMaxKm, units) ?? yMaxKm;

    return { points: strikes, yMaxKm, yMaxDisplay, energyMin, energyMax, xStart, xEnd };
  }, [data, hours, units]);

  const plotW = Math.max(50, width - PAD_LEFT - PAD_RIGHT);
  const plotH = height - PAD_TOP - PAD_BOTTOM;

  const xScale = useCallback((t: number) => {
    return PAD_LEFT + ((t - xStart) / (xEnd - xStart)) * plotW;
  }, [xStart, xEnd, plotW]);

  const yScale = useCallback((km: number) => {
    // closer = bottom (intuitive: storm is "near the ground")
    return PAD_TOP + plotH - (km / yMaxKm) * plotH;
  }, [yMaxKm, plotH]);

  // Y-axis ticks (in display units, with km equivalent for plotting)
  const yTicks = useMemo(() => {
    const target = compact ? 3 : 5;
    const step = niceStep(yMaxDisplay / target);
    const decimals = step >= 1 ? 0 : Math.min(2, Math.ceil(-Math.log10(step)));
    const kmPerDisplay = yMaxDisplay > 0 ? yMaxKm / yMaxDisplay : 1;
    const ticks: { display: number; km: number; label: string }[] = [];
    for (let v = 0; v <= yMaxDisplay + 1e-6; v += step) {
      ticks.push({ display: v, km: v * kmPerDisplay, label: v.toFixed(decimals) });
    }
    return ticks;
  }, [yMaxDisplay, yMaxKm, compact]);

  // X-axis ticks
  const xTicks = useMemo(() => {
    const target = compact ? 4 : 6;
    const totalMs = xEnd - xStart;
    const step = totalMs / target;
    const ticks: number[] = [];
    for (let i = 0; i <= target; i++) ticks.push(xStart + i * step);
    return ticks;
  }, [xStart, xEnd, compact]);

  if (error) return <div className="error">Failed to load strike data</div>;

  return (
    <div ref={containerRef} className="chart-container scatter-container">
      <svg width={width} height={height} style={{ display: 'block' }}>
        {/* grid lines */}
        {yTicks.map((t, i) => {
          const y = yScale(t.km);
          return (
            <line
              key={`gy${i}`}
              x1={PAD_LEFT} x2={PAD_LEFT + plotW}
              y1={y} y2={y}
              stroke="#1e1e30" strokeWidth={1}
            />
          );
        })}

        {/* y-axis labels */}
        {yTicks.map((t, i) => {
          const y = yScale(t.km);
          return (
            <text
              key={`yl${i}`}
              x={PAD_LEFT - 6} y={y + 3}
              textAnchor="end"
              fontSize={compact ? 9 : 10}
              fill="#505060"
            >
              {t.label}
            </text>
          );
        })}
        <text
          x={PAD_LEFT - 6} y={PAD_TOP - 2}
          textAnchor="end" fontSize={compact ? 8 : 9}
          fill="#606070"
        >
          {distanceUnit(units)}
        </text>

        {/* x-axis labels */}
        {xTicks.map((t, i) => (
          <text
            key={`xl${i}`}
            x={xScale(t)} y={height - 10}
            textAnchor={i === 0 ? 'start' : i === xTicks.length - 1 ? 'end' : 'middle'}
            fontSize={compact ? 9 : 10}
            fill="#505060"
          >
            {fmtTimeAxis(new Date(t), hours)}
          </text>
        ))}

        {/* "now" marker */}
        <line
          x1={xScale(xEnd)} x2={xScale(xEnd)}
          y1={PAD_TOP} y2={PAD_TOP + plotH}
          stroke="#2a2a44" strokeWidth={1} strokeDasharray="2,3"
        />

        {/* strikes */}
        {points.map((s, i) => {
          const t = new Date(s.time).getTime();
          if (t < xStart || t > xEnd) return null;
          const cx = xScale(t);
          const cy = yScale(s.distance_km);
          const intensity = energyIntensity(s.energy, energyMin, energyMax);
          const r = 2.5 + intensity * (compact ? 4 : 6);
          const color = intensityColor(intensity);
          return (
            <circle
              key={i}
              cx={cx} cy={cy} r={r}
              fill={color}
              fillOpacity={0.75}
              stroke={color}
              strokeOpacity={0.95}
              onMouseEnter={() => setHover({ strike: s, x: cx, y: cy })}
              onMouseLeave={() => setHover(h => h?.strike === s ? null : h)}
              style={{ cursor: 'pointer' }}
            />
          );
        })}

        {points.length === 0 && (
          <text
            x={PAD_LEFT + plotW / 2} y={PAD_TOP + plotH / 2}
            textAnchor="middle"
            fontSize={11} fill="#505060"
          >
            No strikes in selected range
          </text>
        )}
      </svg>

      {hover && (
        <div
          className="scatter-tooltip"
          style={{
            left: Math.min(width - 160, Math.max(0, hover.x + 8)),
            top: Math.max(0, hover.y - 48),
          }}
        >
          <div>{fmtTooltipTime(new Date(hover.strike.time))}</div>
          <div>
            <strong>{(distanceConvert(hover.strike.distance_km, units) ?? 0).toFixed(1)} {distanceUnit(units)}</strong>
            {' · '}
            energy {hover.strike.energy.toLocaleString()}
          </div>
        </div>
      )}
    </div>
  );
}

function niceStep(raw: number): number {
  if (raw <= 0) return 1;
  const pow = Math.pow(10, Math.floor(Math.log10(raw)));
  const n = raw / pow;
  if (n < 1.5) return 1 * pow;
  if (n < 3) return 2 * pow;
  if (n < 7) return 5 * pow;
  return 10 * pow;
}
