import { median, boxPlot } from '../reports/statistics.mjs';

function barValue(values) {
  const value = median([...values].sort((a, b) => a - b));
  // Go math.Round rounds ties away from zero, unlike Math.round for negatives.
  return Math.sign(value) * Math.round(Math.abs(value) * 1e5) / 1e5;
}

export function reportCharts(report) {
  if (report.version !== 1) throw new Error(`Unsupported report version: ${report.version}`);
  if (!['buildkit', 'buildx'].includes(report.project)) throw new Error('Unknown report project');
  const candidates = report.candidates[report.project] || [];
  return [...report.benchmarks].sort((a, b) => a.name < b.name ? -1 : a.name > b.name ? 1 : 0)
    .flatMap(benchmark => {
      const metrics = new Map();
      for (const sample of benchmark.samples) {
        for (const [unit, value] of Object.entries(sample.values)) {
          if (!Number.isFinite(value)) throw new Error(`Invalid ${unit} value in ${benchmark.name}`);
          if (!benchmark.metrics[unit]) throw new Error(`Unknown metric: ${unit}`);
          if (!metrics.has(unit)) metrics.set(unit, new Map());
          const refs = metrics.get(unit);
          const ref = sample[report.project];
          if (!refs.has(ref)) refs.set(ref, []);
          refs.get(ref).push(value);
        }
      }
      const groups = [...metrics.keys()].sort().map(unit => [unit]);
      if (['duration', 'alloc'].every(unit => metrics.has(unit) && benchmark.metrics[unit].chart === 'boxplot')) {
        groups.splice(0, groups.length, ['duration', 'alloc'], ...groups.filter(([unit]) => !['duration', 'alloc'].includes(unit)));
      }
      return groups.map(units => {
        const first = metrics.get(units[0]);
        const refs = (candidates.length ? candidates.map(candidate => candidate.name) : [...first.keys()].sort())
          .filter(ref => units.every(unit => metrics.get(unit).has(ref)));
        const series = units.map((unit, index) => {
          const config = benchmark.metrics[unit];
          if (!['bar', 'boxplot'].includes(config.chart)) throw new Error(`Unknown chart type: ${config.chart}`);
          return {
            name: config.description,
            type: config.chart,
            yAxisIndex: index,
            data: refs.map(ref => ({ value: (config.chart === 'bar' ? barValue : boxPlot)(metrics.get(unit).get(ref)) })),
            ...(config.chart === 'bar' && config.average ? {
              markLine: { data: [{ name: 'Avg', type: 'average', lineStyle: { color: '#91cc75', width: 2 } }] },
            } : {}),
          };
        });
        return {
          id: `${benchmark.name}/${units.join('+')}`,
          option: {
            color: ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc'],
            title: { text: benchmark.description, subtext: benchmark.name },
            legend: { top: 60 },
            tooltip: { trigger: 'item' },
            grid: { top: 110, left: 90, right: 90, bottom: 100 },
            dataZoom: [{ type: 'slider', bottom: 15 }],
            toolbox: { right: 10, feature: { dataZoom: {}, restore: {}, saveAsImage: {} } },
            xAxis: { type: 'category', data: refs },
            yAxis: units.map((unit, index) => ({
              type: 'value', name: benchmark.metrics[unit].description,
              nameLocation: 'middle', nameGap: 60,
              position: index === 0 ? 'left' : 'right', alignTicks: units.length > 1,
            })),
            series,
          },
        };
      });
    });
}
