import { boxPlot } from './statistics.mjs';

export function candidateRows(report, benchmark, metric) {
  if (!benchmark || !metric) return [];
  const values = new Map();
  for (const sample of benchmark.samples) {
    const value = sample.values[metric];
    if (!Number.isFinite(value)) continue;
    const ref = sample[report.project];
    if (!values.has(ref)) values.set(ref, []);
    values.get(ref).push(value);
  }
  const candidates = report.candidates[report.project] || [];
  const refs = candidates.length ? candidates.map((c) => c.name) : [...values.keys()].sort();
  return refs
    .filter((ref) => values.has(ref))
    .map((ref) => {
      const samples = values.get(ref);
      const [min, q1, median, q3, max] = boxPlot(samples);
      return {
        ref,
        min,
        q1,
        median,
        q3,
        max,
        count: samples.length,
        sha: candidates.find((c) => c.name === ref)?.sha || '',
      };
    });
}

export function percentChange(value, baseline) {
  if (!Number.isFinite(value) || !Number.isFinite(baseline) || baseline === 0) return null;
  return ((value - baseline) / Math.abs(baseline)) * 100;
}

export function sourceProfiles(report, benchmark, ref) {
  const samples = (benchmark?.samples || []).filter((sample) => !ref || sample[report.project] === ref);
  return [...new Set(samples.map((sample) => sample.source_id))].map((id) => ({
    id,
    name: report.sources?.[id]?.name || samples.find((s) => s.source_id === id)?.source || 'Unknown source',
    execution: report.sources?.[id]?.execution,
    samples: samples.filter((sample) => sample.source_id === id),
  }));
}

function unique(values) {
  return [
    ...new Set(
      values.map((value) =>
        value === undefined || value === '' || value === null ? 'Not recorded' : String(value)
      )
    ),
  ].sort();
}

function contextFields(prefix, objects) {
  const keys = [...new Set(objects.flatMap((object) => Object.keys(object || {})))].sort();
  return keys.length
    ? Object.fromEntries(
        keys.map((key) => [
          key === 'suite_revision' ? 'Suite revision' : `${prefix}: ${key}`,
          unique(objects.map((object) => (object?.[key] === '' ? '(empty)' : object?.[key]))),
        ])
      )
    : { [prefix]: ['Not recorded'] };
}

export function comparisonContext(report, benchmark, ref) {
  const samples = benchmark?.samples.filter((sample) => sample[report.project] === ref) || [];
  if (!samples.length) return {};
  const executions = sourceProfiles(report, benchmark, ref).map((source) => source.execution);
  const environments = executions.map((execution) => {
    const env = execution?.environment;
    if (!env) return undefined;
    return Object.fromEntries(Object.entries(env).filter(([key]) => !key.startsWith('GITHUB_')));
  });
  const counterpart = report.project === 'buildkit' ? 'buildx' : 'buildkit';
  return {
    CPU: unique(samples.map((s) => s.cpu)),
    'OS / architecture': unique(
      samples.map((s) => (s.os && s.architecture ? `${s.os} / ${s.architecture}` : undefined))
    ),
    'Go runtime': unique(executions.map((e) => e?.go_version)),
    'CPU count / GOMAXPROCS': unique(
      executions.map((e) => (e ? `${e.num_cpu} / ${e.gomaxprocs}` : undefined))
    ),
    'Iterations per sample': unique(samples.map((s) => s.iterations)),
    ...contextFields(
      'Setting',
      executions.map((e) => e?.settings)
    ),
    ...contextFields('Environment', environments),
    'Counterpart binaries': unique(
      samples.map((s) => {
        const binaries = report.sources?.[s.source_id]?.execution?.binaries;
        const matches = binaries?.filter((b) => b.project === counterpart && b.ref === s[counterpart]);
        return matches?.length
          ? matches
              .map((b) => `${b.name}:${b.sha256}`)
              .sort()
              .join(', ')
          : undefined;
      })
    ),
  };
}

export function compareContexts(current, other) {
  const fields = [...new Set([...Object.keys(current), ...Object.keys(other)])];
  const missing = fields.filter(
    (key) =>
      !current[key] ||
      !other[key] ||
      current[key].includes('Not recorded') ||
      other[key].includes('Not recorded')
  );
  const changed = fields.filter(
    (key) => !missing.includes(key) && JSON.stringify(current[key]) !== JSON.stringify(other[key])
  );
  return {
    missing,
    changed,
    label: changed.length
      ? 'Recorded differences'
      : missing.length || !fields.length
      ? 'Metadata incomplete'
      : 'Recorded fields match',
  };
}

// Group only this benchmark family; keep original reports and samples intact.
export function benchmarkNavigation(benchmarks) {
  const entries = [];
  const groups = new Map();
  for (const benchmark of benchmarks) {
    const match = benchmark.name.match(/^(.*\/BenchmarkBuildHighParallelization)(16|32|64|128)x$/);
    if (!match) {
      entries.push(benchmark);
      continue;
    }
    let group = groups.get(match[1]);
    if (!group) {
      group = { name: match[1], description: 'Build high parallelization', variants: [] };
      groups.set(match[1], group);
      entries.push(group);
    }
    group.variants.push({ parallelism: Number(match[2]), benchmark });
  }
  for (const group of groups.values()) group.variants.sort((a, b) => a.parallelism - b.parallelism);
  return entries;
}

export function parallelizationRows(report, variants, metric, unit) {
  const levels = variants.map(({ benchmark }) =>
    benchmark.metrics[metric]?.unit === unit ? candidateRows(report, benchmark, metric) : []
  );
  const refs = new Set(levels.flatMap((rows) => rows.map((row) => row.ref)));
  const candidates = report.candidates[report.project] || [];
  const order = candidates.length ? candidates.map((candidate) => candidate.name) : [...refs].sort();
  return order
    .filter((ref) => refs.has(ref))
    .map((ref) => ({ ref, points: levels.map((rows) => rows.find((row) => row.ref === ref) || null) }));
}
