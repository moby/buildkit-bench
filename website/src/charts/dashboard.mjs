import { formatValue } from '../reports/format.mjs';

export function chartPalette(theme) {
  return theme === 'dark'
    ? {
        text: '#a5b2c8',
        grid: '#293247',
        accent: '#70a8ff',
        fill: '#233d62',
        background: '#171e2c',
        border: '#38445b',
      }
    : {
        text: '#56657c',
        grid: '#e5eaf1',
        accent: '#326ce5',
        fill: '#e3edff',
        background: '#ffffff',
        border: '#cbd5e4',
      };
}

export function dashboardChart(option, theme, baseline, mode = 'distribution', rows = [], unit = '') {
  const palette = chartPalette(theme);
  return {
    ...option,
    backgroundColor: 'transparent',
    title: { show: false },
    legend: { show: false },
    textStyle: { fontFamily: 'Inter, Segoe UI, sans-serif', color: palette.text },
    color: [palette.accent, '#5fcca6'],
    animationDuration: 250,
    grid: { top: 36, left: 72, right: 28, bottom: 90 },
    tooltip: {
      trigger: 'item',
      valueFormatter: (value) =>
        Array.isArray(value) ? value.map((item) => formatValue(item, unit)) : formatValue(value, unit),
      backgroundColor: palette.background,
      borderColor: palette.border,
      textStyle: { color: palette.text },
    },
    toolbox: { show: false },
    xAxis: {
      ...option.xAxis,
      axisLine: { lineStyle: { color: palette.grid } },
      axisTick: { show: false },
      axisLabel: { color: palette.text, fontSize: 11, rotate: option.xAxis.data.length > 12 ? 30 : 0 },
    },
    yAxis: option.yAxis.map((axis) => ({
      ...axis,
      name: '',
      alignTicks: false,
      axisLabel: {
        color: palette.text,
        formatter: (value) => formatValue(value, unit),
      },
      splitLine: { lineStyle: { color: palette.grid, type: 'dashed' } },
    })),
    dataZoom: [
      {
        type: 'slider',
        bottom: 4,
        height: 20,
        borderColor: 'transparent',
        backgroundColor: palette.grid,
        fillerColor: `${palette.accent}22`,
        textStyle: { color: palette.text },
        handleStyle: { color: palette.accent, borderColor: palette.accent },
      },
    ],
    series: option.series.map((series) => ({
      ...series,
      ...(mode === 'medians'
        ? {
            type: 'line',
            data: rows.map((row) => ({ value: row.median })),
            smooth: false,
            symbol: 'circle',
            symbolSize: 7,
            lineStyle: { width: 2 },
            areaStyle: { opacity: 0.07 },
          }
        : {}),
      itemStyle: {
        color: mode === 'medians' ? palette.accent : palette.fill,
        borderColor: palette.accent,
        borderWidth: 1.5,
      },
      ...(baseline
        ? {
            markLine: {
              symbol: 'none',
              silent: true,
              label: { show: false },
              lineStyle: { color: '#a88aef', type: 'dashed', width: 1.5 },
              data: [
                ...(series.markLine?.data || []),
                { name: `Baseline: ${baseline.ref}`, yAxis: baseline.median },
              ],
            },
          }
        : {}),
    })),
  };
}
