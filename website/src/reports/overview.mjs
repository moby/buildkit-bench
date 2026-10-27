import {
  candidateRows,
  percentChange,
  comparisonContext,
  compareContexts,
  benchmarkNavigation,
} from './benchmarks.mjs';

// These are the report's known lower-is-better metrics. Custom metrics have no
// direction contract, so they are inspected for coverage but never ranked.
const directions = { duration: 's', alloc: 'bytes', bytes: 'bytes' };

// Select by numeric version, not candidate order or commit date. Prereleases
// are not stable baselines, and no external release lookup is needed.
export function latestStableCandidate(candidates) {
  return candidates
    .map((candidate) => ({
      candidate,
      version: /^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\+[0-9A-Za-z.-]+)?$/.exec(candidate.name),
    }))
    .filter((entry) => entry.version)
    .sort(
      (a, b) =>
        Number(b.version[1]) - Number(a.version[1]) ||
        Number(b.version[2]) - Number(a.version[2]) ||
        Number(b.version[3]) - Number(a.version[3])
    )[0]?.candidate;
}

export function reportOverview(report, threshold = 10) {
  const candidates = report.candidates[report.project] || [];
  const baseline = latestStableCandidate(candidates);
  const target = candidates.find((candidate) => candidate.name === 'master');
  const findings = [];
  const issues = [];
  let comparisons = 0;
  let covered = 0;
  let expected = 0;
  const issue = (benchmark, metric, ref, message) => issues.push({ benchmark, metric, ref, message });
  if (!baseline)
    issue(
      null,
      '',
      '',
      'No stable release is recorded in this report; baseline comparisons are unavailable.'
    );

  if (!target)
    issue(
      null,
      '',
      'master',
      'Master is not recorded in this report; regression comparisons are unavailable.'
    );

  for (const benchmark of report.benchmarks) {
    const contexts = new Map();
    const context = (ref) => {
      if (!contexts.has(ref)) contexts.set(ref, comparisonContext(report, benchmark, ref));
      return contexts.get(ref);
    };
    for (const [metric, config] of Object.entries(benchmark.metrics)) {
      const rows = new Map(candidateRows(report, benchmark, metric).map((row) => [row.ref, row]));
      const invalid = new Set();
      for (const candidate of candidates) {
        expected++;
        const row = rows.get(candidate.name);
        if (!row) {
          issue(benchmark, metric, candidate.name, 'No measurements recorded.');
          continue;
        }
        covered++;
        const sampleCount = benchmark.samples.filter(
          (sample) => sample[report.project] === candidate.name
        ).length;
        if (row.count < sampleCount)
          issue(
            benchmark,
            metric,
            candidate.name,
            `Metric absent from ${sampleCount - row.count} of ${sampleCount} samples.`
          );
        if (directions[metric] === config.unit && row.min < 0) {
          invalid.add(candidate.name);
          issue(benchmark, metric, candidate.name, 'Negative measurements for a non-negative metric.');
        } else if (metric === 'duration' && row.min === 0) {
          issue(benchmark, metric, candidate.name, 'Zero duration recorded; inspect the raw measurements.');
        }
        if (config.chart === 'boxplot' && row.count < 3) {
          issue(
            benchmark,
            metric,
            candidate.name,
            `Only ${row.count} sample(s); limited evidence of repeatability.`
          );
        }
        if (row.median > 0 && (row.max - row.min) / row.median > 0.5) {
          issue(
            benchmark,
            metric,
            candidate.name,
            'Observed range exceeds 50% of the median; measurements are variable.'
          );
        }
      }
      if (directions[metric] !== config.unit) continue;
      // Never substitute an older release when this baseline lacks measurements.
      const before = baseline && rows.get(baseline.name);
      const after = target && rows.get(target.name);
      if (!before || !after || invalid.has(before.ref) || invalid.has(after.ref)) continue;
      comparisons++;
      const delta = percentChange(after.median, before.median);
      if (after.median <= before.median || (delta !== null && delta < threshold)) continue;
      const comparison = compareContexts(context(after.ref), context(before.ref));
      const sameRevision = !!before.sha && before.sha === after.sha;
      const separated = after.min > before.max;
      findings.push({
        benchmark,
        metric,
        config,
        before,
        after,
        delta,
        comparison,
        sameRevision,
        separated,
        repeated: before.count >= 3 && after.count >= 3,
        // A triage heuristic, not a significance test or a causal conclusion.
        priority:
          !sameRevision &&
          before.median > 0 &&
          separated &&
          before.count >= 3 &&
          after.count >= 3 &&
          !comparison.changed.length,
      });
    }
  }
  findings.sort(
    (a, b) => Number(b.priority) - Number(a.priority) || (b.delta ?? Infinity) - (a.delta ?? Infinity)
  );
  return {
    baseline,
    target,
    findings,
    issues,
    comparisons,
    covered,
    expected,
    candidates: candidates.length,
  };
}

export function evidenceLocation(report, run, benchmark, metric, baseline, candidate) {
  const entry = benchmarkNavigation(report.benchmarks).find(
    (entry) =>
      entry.name === benchmark.name || entry.variants?.some((v) => v.benchmark.name === benchmark.name)
  );
  return {
    name: 'ResultView',
    params: { result: run },
    query: {
      view: 'benchmarks',
      project: report.project,
      benchmark: entry.name,
      variant: entry.variants ? benchmark.name : undefined,
      metric,
      baseline,
      candidate,
      chart: 'distribution',
    },
  };
}

export function revisionComparison(project, before, after) {
  if (![before, after].every((sha) => /^[0-9a-f]{40}$/i.test(sha || '')) || before === after) return null;
  const repository = { buildkit: 'moby/buildkit', buildx: 'docker/buildx' }[project];
  return repository ? `https://github.com/${repository}/compare/${before}...${after}` : null;
}
