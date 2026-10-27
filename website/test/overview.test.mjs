import assert from 'node:assert/strict';
import test from 'node:test';
import {
  reportOverview,
  evidenceLocation,
  revisionComparison,
  latestStableCandidate,
} from '../src/reports/overview.mjs';

function fixture(
  values = [
    [10, 11, 12],
    [20, 21, 22],
  ],
  project = 'buildkit'
) {
  return {
    version: 1,
    project,
    candidates: {
      [project]: values.map((_, i) => ({
        name: i === 0 ? 'v1.0.0' : i === values.length - 1 ? 'master' : `ref-${i}`,
        sha: String(i + 1).repeat(40),
      })),
    },
    benchmarks: [
      {
        name: 'BenchmarkBuild/BenchmarkBuildSimple',
        description: 'Simple build',
        metrics: { duration: { unit: 's', chart: 'boxplot' } },
        samples: values.flatMap((rows, i) =>
          rows.map((value) => ({
            [project]: i === 0 ? 'v1.0.0' : i === values.length - 1 ? 'master' : `ref-${i}`,
            cpu: 'cpu-a',
            os: 'linux',
            architecture: 'amd64',
            iterations: 1,
            values: { duration: value },
          }))
        ),
      },
    ],
  };
}

test('ranks repeated separated increases with raw evidence, not a significance claim', () => {
  const report = fixture();
  const snapshot = JSON.stringify(report);
  const result = reportOverview(report);
  assert.equal(result.comparisons, 1);
  assert.equal(result.covered, 2);
  assert.equal(result.expected, 2);
  assert.equal(result.findings.length, 1);
  const finding = result.findings[0];
  assert.equal(finding.before.median, 11);
  assert.equal(finding.after.median, 21);
  assert.equal(finding.before.count, 3);
  assert.ok(Math.abs(finding.delta - 1000 / 11) < 1e-10);
  assert.equal(finding.priority, true);
  assert.ok(finding.comparison.missing.includes('Go runtime'));
  assert.equal(JSON.stringify(report), snapshot);
});

test('latest stable selection uses numeric versions and excludes prereleases and branch names', () => {
  const candidates = ['master', 'v0.9.9', 'v0.10.2', 'v0.10.10', 'v1.0.0-rc.1', 'pr-123', '2026-10-22'].map(
    (name) => ({ name })
  );
  assert.equal(latestStableCandidate(candidates).name, 'v0.10.10');
  assert.equal(latestStableCandidate([...candidates, { name: 'v1.0.0' }]).name, 'v1.0.0');
  assert.equal(latestStableCandidate([{ name: 'v1.0.0-beta.1' }, { name: 'master' }]), undefined);
});

test('every comparison uses the highest stable release even when it is not first', () => {
  const report = fixture([
    [30, 30, 30],
    [10, 10, 10],
    [20, 20, 20],
  ]);
  report.candidates.buildkit[1].name = 'v1.1.0';
  report.benchmarks[0].samples
    .filter((s) => s.buildkit === 'ref-1')
    .forEach((s) => {
      s.buildkit = 'v1.1.0';
    });
  const result = reportOverview(report);
  assert.equal(result.baseline.name, 'v1.1.0');
  assert.equal(result.comparisons, 1);
  assert.equal(result.findings.length, 1);
  assert.ok(result.findings.every((finding) => finding.before.ref === 'v1.1.0'));
});

test('missing stable baseline measurements do not silently use a different candidate', () => {
  const report = fixture([[], [10, 10, 10], [20, 20, 20]]);
  const result = reportOverview(report);
  assert.equal(result.baseline.name, 'v1.0.0');
  assert.equal(result.comparisons, 0);
  assert.equal(result.findings.length, 0);
  assert.ok(
    result.issues.some((issue) => issue.ref === 'v1.0.0' && issue.message.includes('No measurements'))
  );
});

test('threshold, decreases, and unchanged medians do not produce false increases', () => {
  assert.equal(
    reportOverview(
      fixture([
        [100, 100, 100],
        [109, 109, 109],
      ])
    ).findings.length,
    0
  );
  assert.equal(
    reportOverview(
      fixture([
        [100, 100, 100],
        [110, 110, 110],
      ])
    ).findings.length,
    1
  );
  assert.equal(
    reportOverview(
      fixture([
        [100, 100, 100],
        [110, 110, 110],
      ]),
      20
    ).findings.length,
    0
  );
  assert.equal(
    reportOverview(
      fixture([
        [20, 21, 22],
        [10, 11, 12],
      ])
    ).findings.length,
    0
  );
  assert.equal(
    reportOverview(
      fixture([
        [10, 10, 10],
        [10, 10, 10],
      ])
    ).findings.length,
    0
  );
});

test('only master is compared even when another candidate has a much larger increase', () => {
  const report = fixture([
    [10, 10, 10],
    [100, 100, 100],
    [20, 20, 20],
  ]);
  const result = reportOverview(report);
  assert.equal(result.target.name, 'master');
  assert.equal(result.comparisons, 1);
  assert.equal(result.findings.length, 1);
  assert.equal(result.findings[0].after.ref, 'master');
  assert.equal(result.findings[0].delta, 100);
});

test('missing master or its measurements never substitutes another candidate', () => {
  const report = fixture();
  report.candidates.buildkit[1].name = 'pr-123';
  report.benchmarks[0].samples
    .filter((s) => s.buildkit === 'master')
    .forEach((s) => {
      s.buildkit = 'pr-123';
    });
  const result = reportOverview(report);
  assert.equal(result.target, undefined);
  assert.equal(result.comparisons, 0);
  assert.equal(result.findings.length, 0);
  assert.ok(result.issues.some((issue) => issue.message.includes('Master is not recorded')));
  const missingSamples = reportOverview(fixture([[10, 10, 10], [100, 100, 100], []]));
  assert.equal(missingSamples.comparisons, 0);
  assert.equal(missingSamples.findings.length, 0);
  assert.ok(
    missingSamples.issues.some((issue) => issue.ref === 'master' && issue.message.includes('No measurements'))
  );
});

test('missing candidate measurements do not prevent other stable-baseline comparisons', () => {
  const result = reportOverview(fixture([[10, 11, 12], [], [20, 21, 22]]));
  assert.equal(result.findings.length, 1);
  assert.equal(result.comparisons, 1);
  assert.equal(result.covered, 2);
  assert.equal(result.expected, 3);
  assert.equal(result.issues[0].ref, 'ref-1');
  assert.match(result.issues[0].message, /No measurements/);
});

test('same SHA, changed hardware, sparse samples, and overlapping ranges lower priority', () => {
  const same = fixture();
  same.candidates.buildkit[1].sha = same.candidates.buildkit[0].sha;
  assert.equal(reportOverview(same).findings[0].sameRevision, true);
  assert.equal(reportOverview(same).findings[0].priority, false);
  const changed = fixture();
  changed.benchmarks[0].samples[3].cpu = 'cpu-b';
  const finding = reportOverview(changed).findings[0];
  assert.deepEqual(finding.comparison.changed, ['CPU']);
  assert.equal(finding.priority, false);
  assert.equal(reportOverview(fixture([[10], [20]])).findings[0].priority, false);
  const noisy = reportOverview(
    fixture([
      [1, 10, 30],
      [9, 20, 40],
    ])
  );
  assert.equal(noisy.findings[0].separated, false);
  assert.equal(noisy.findings[0].priority, false);
  assert.ok(noisy.issues.some((issue) => issue.message.includes('variable')));
});

test('zero baselines have no invented percentage and negative values cannot rank', () => {
  const zero = reportOverview(
    fixture([
      [0, 0, 0],
      [1, 1, 1],
    ])
  );
  assert.equal(zero.findings[0].delta, null);
  assert.equal(zero.findings[0].priority, false);
  assert.ok(zero.issues.some((issue) => issue.message.includes('Zero duration')));
  const negative = reportOverview(
    fixture([
      [-1, -1, -1],
      [1, 1, 1],
    ])
  );
  assert.equal(negative.findings.length, 0);
  assert.ok(negative.issues.some((issue) => issue.message.includes('Negative')));
});

test('unknown directions and mismatched units are excluded from increase ranking', () => {
  const report = fixture();
  report.benchmarks[0].metrics.duration.unit = 'custom';
  assert.equal(reportOverview(report).comparisons, 0);
  report.benchmarks[0].metrics = { throughput: { unit: 'bytes/s', chart: 'bar' } };
  report.benchmarks[0].samples.forEach((sample) => {
    sample.values.throughput = sample.values.duration;
  });
  assert.equal(reportOverview(report).findings.length, 0);
});

test('partial metrics and absent stable release are visible', () => {
  const report = fixture();
  delete report.benchmarks[0].samples[0].values.duration;
  assert.ok(reportOverview(report).issues.some((issue) => issue.message.includes('Metric absent')));
  report.candidates = {};
  const result = reportOverview(report);
  assert.equal(result.comparisons, 0);
  assert.match(result.issues[0].message, /No stable release/);
});

test('Buildx comparisons use Buildx refs and evidence links focus grouped variants', () => {
  const report = fixture(undefined, 'buildx');
  assert.equal(reportOverview(report).findings[0].after.ref, 'master');
  const benchmark = report.benchmarks[0];
  benchmark.name = 'BenchmarkBuild/BenchmarkBuildHighParallelization128x';
  const location = evidenceLocation(report, 'run-a', benchmark, 'duration', 'ref-0', 'ref-1');
  assert.equal(location.params.result, 'run-a');
  assert.equal(location.query.project, 'buildx');
  assert.equal(location.query.benchmark, 'BenchmarkBuild/BenchmarkBuildHighParallelization');
  assert.equal(location.query.variant, benchmark.name);
  assert.equal(location.query.baseline, 'ref-0');
  assert.equal(location.query.candidate, 'ref-1');
  assert.equal(location.query.view, 'benchmarks');
});

test('revision comparison links require distinct recorded full SHAs', () => {
  const a = 'a'.repeat(40),
    b = 'b'.repeat(40);
  assert.equal(revisionComparison('buildkit', a, b), `https://github.com/moby/buildkit/compare/${a}...${b}`);
  assert.equal(revisionComparison('buildx', a, b), `https://github.com/docker/buildx/compare/${a}...${b}`);
  assert.equal(revisionComparison('buildkit', a, a), null);
  assert.equal(revisionComparison('buildkit', '../bad', b), null);
  assert.equal(revisionComparison('other', a, b), null);
});
