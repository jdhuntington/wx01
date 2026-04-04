import { useEffect, useRef, useCallback } from 'react';
import { createChart, LineSeries, HistogramSeries } from 'lightweight-charts';
import type { IChartApi, UTCTimestamp } from 'lightweight-charts';
import { useFetch } from '../hooks';

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
  compact?: boolean;
  version: number;
}

export default function TimeSeriesChart({ title, fetcher, series, range_, yFormat, transform, compact, version }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);

  const boundFetcher = useCallback(() => fetcher(range_), [fetcher, range_]);
  const { data, error } = useFetch(boundFetcher, version);

  useEffect(() => {
    if (!containerRef.current) return;

    const chartHeight = compact ? 100 : 220;
    const chart = createChart(containerRef.current, {
      layout: {
        background: { color: '#141424' },
        textColor: '#505060',
        fontSize: compact ? 9 : 11,
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
      height: chartHeight,
      timeScale: {
        timeVisible: !compact,
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
  }, [yFormat, compact]);

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

    // Show full requested range, not just the data we have.
    // setVisibleRange clips to data boundaries, so we use setVisibleLogicalRange.
    // Logical index 0 = first data point, negative = before data.
    // We calculate how many data points would fill the full range and offset accordingly.
    const rangeHours = parseInt(range_) || 24;
    const rangeSeconds = rangeHours * 3600;
    const nowSeconds = Date.now() / 1000;

    // Find the earliest data point timestamp across all series
    let earliestDataTime = nowSeconds;
    for (const s of series) {
      const first = data.find((d: any) => d[s.key] !== null && d[s.key] !== undefined);
      if (first) {
        const t = new Date(first.bucket).getTime() / 1000;
        if (t < earliestDataTime) earliestDataTime = t;
      }
    }

    // Average interval between data points (for logical index math)
    const dataCount = data.length;
    const lastTime = new Date(data[data.length - 1].bucket).getTime() / 1000;
    const firstTime = new Date(data[0].bucket).getTime() / 1000;
    const avgInterval = dataCount > 1 ? (lastTime - firstTime) / (dataCount - 1) : 1;

    // How many logical indices before the first point does the range start?
    const rangeStart = nowSeconds - rangeSeconds;
    const logicalFrom = (rangeStart - firstTime) / avgInterval;
    const logicalTo = (nowSeconds - firstTime) / avgInterval;

    chart.timeScale().setVisibleLogicalRange({ from: logicalFrom, to: logicalTo });

    return () => {
      for (const s of seriesRefs) {
        try { chart.removeSeries(s); } catch {}
      }
    };
  }, [data, series, yFormat, transform, range_]);

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
