import { useEffect, useRef, useCallback } from 'react';
import { createChart, LineSeries, HistogramSeries } from 'lightweight-charts';
import type { IChartApi, UTCTimestamp } from 'lightweight-charts';
import { usePolling } from '../hooks';

interface SeriesDef {
  name: string;
  key: string;
  color: string;
  type?: 'line' | 'histogram';
}

interface Props {
  title: string;
  fetcher: (range_: string) => Promise<any[]>;
  series: SeriesDef[];
  range_: string;
  yFormat?: (v: number) => string;
  transform?: (v: number) => number;
}

export default function TimeSeriesChart({ title, fetcher, series, range_, yFormat, transform }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);

  const boundFetcher = useCallback(() => fetcher(range_), [fetcher, range_]);
  const { data, error } = usePolling(boundFetcher, 30_000);

  useEffect(() => {
    if (!containerRef.current) return;

    const chart = createChart(containerRef.current, {
      layout: {
        background: { color: '#141424' },
        textColor: '#505060',
        fontSize: 11,
        attributionLogo: false,
      },
      grid: {
        vertLines: { color: '#1e1e30' },
        horzLines: { color: '#1e1e30' },
      },
      crosshair: {
        vertLine: { color: '#2a2a44', width: 1, labelBackgroundColor: '#2a2a44' },
        horzLine: { color: '#2a2a44', width: 1, labelBackgroundColor: '#2a2a44' },
      },
      rightPriceScale: {
        borderColor: '#1e1e30',
      },
      width: containerRef.current.clientWidth,
      height: 220,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: '#1e1e30',
      },
      localization: {
        priceFormatter: yFormat,
      },
    });

    chartRef.current = chart;

    const ro = new ResizeObserver(entries => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width });
      }
    });
    ro.observe(containerRef.current);

    return () => {
      ro.disconnect();
      chart.remove();
      chartRef.current = null;
    };
  }, [yFormat]);

  useEffect(() => {
    const chart = chartRef.current;
    if (!chart || !data || data.length === 0) return;

    chart.applyOptions({
      localization: { priceFormatter: yFormat },
    });

    const seriesRefs: any[] = [];

    for (const s of series) {
      const seriesData = data
        .filter((d: any) => d[s.key] !== null && d[s.key] !== undefined)
        .map((d: any) => {
          const raw = d[s.key] as number;
          return {
            time: Math.floor(new Date(d.bucket).getTime() / 1000) as UTCTimestamp,
            value: transform ? transform(raw) : raw,
          };
        });

      if (seriesData.length === 0) continue;

      if (s.type === 'histogram') {
        const histSeries = chart.addSeries(HistogramSeries, {
          color: s.color,
          priceFormat: { type: 'price', precision: 2, minMove: 0.01 },
        });
        histSeries.setData(seriesData);
        seriesRefs.push(histSeries);
      } else {
        const lineSeries = chart.addSeries(LineSeries, {
          color: s.color,
          lineWidth: 2,
          lineType: 2, // curved
          priceFormat: { type: 'price', precision: 1, minMove: 0.1 },
        });
        lineSeries.setData(seriesData);
        seriesRefs.push(lineSeries);
      }
    }

    chart.timeScale().fitContent();

    return () => {
      for (const s of seriesRefs) {
        try { chart.removeSeries(s); } catch {}
      }
    };
  }, [data, series, yFormat, transform]);

  const showLegend = series.length > 1;

  return (
    <div className="card chart-card">
      <div className="chart-header">
        <div className="chart-title">{title}</div>
        {showLegend && (
          <div className="chart-legend">
            {series.map(s => (
              <div key={s.key} className="legend-item">
                <span className="legend-swatch" style={{ backgroundColor: s.color }} />
                {s.name}
              </div>
            ))}
          </div>
        )}
      </div>
      {error && <div className="error">Error loading data</div>}
      <div ref={containerRef} className="chart-container" />
    </div>
  );
}
