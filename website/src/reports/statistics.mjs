export function median(sorted) {
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 ? sorted[middle] : (sorted[middle - 1] + sorted[middle]) / 2;
}

// Match the historical Go renderer and montanaflynn/stats, not ECharts' quartile transform:
// exclude the middle observation from odd halves and use min/max whiskers.
export function boxPlot(values) {
  const sorted = [...values].sort((a, b) => a - b);
  if (sorted.length === 1) return Array(5).fill(sorted[0]);
  const middle = Math.floor(sorted.length / 2);
  return [sorted[0], median(sorted.slice(0, middle)), median(sorted),
    median(sorted.slice(Math.ceil(sorted.length / 2))), sorted[sorted.length - 1]];
}
