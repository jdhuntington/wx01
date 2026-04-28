import { useState } from 'react';
import TimeSeriesChart from './TimeSeriesChart';
import LightningScatter from './LightningScatter';
import { getLightning } from '../api';
import type { UnitSystem } from '../units';

type View = 'histogram' | 'detail';

interface Props {
  range_: string;
  units: UnitSystem;
  compact?: boolean;
  version: number;
}

export default function LightningPanel({ range_, units, compact, version }: Props) {
  const [view, setView] = useState<View>('histogram');

  const toggle = (
    <div className="view-toggle">
      <button
        className={view === 'histogram' ? 'active' : ''}
        onClick={() => setView('histogram')}
      >
        Count
      </button>
      <button
        className={view === 'detail' ? 'active' : ''}
        onClick={() => setView('detail')}
      >
        Detail
      </button>
    </div>
  );

  if (view === 'histogram') {
    return (
      <TimeSeriesChart
        title="Lightning"
        fetcher={getLightning}
        range_={range_}
        series={[{ name: 'Strikes', key: 'strike_count', color: '#f1c40f', type: 'histogram', precision: 0 }]}
        yFormat={(v) => `${v.toFixed(0)} ${v === 1 ? 'strike' : 'strikes'}`}
        compact={compact}
        version={version}
        headerExtra={toggle}
      />
    );
  }

  return (
    <div className="card chart-card">
      <div className="chart-header">
        <div className="chart-title">Lightning · distance vs time</div>
        <div className="chart-header-right">
          <div className="chart-legend">
            <div className="legend-item">
              <span className="legend-swatch" style={{ backgroundColor: '#f1c40f' }} />
              low energy
            </div>
            <div className="legend-item">
              <span className="legend-swatch" style={{ backgroundColor: '#e74c3c' }} />
              high energy
            </div>
          </div>
          {toggle}
        </div>
      </div>
      <LightningScatter range_={range_} units={units} version={version} compact={compact} />
    </div>
  );
}
