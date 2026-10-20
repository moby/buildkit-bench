<template>
  <section class="details-panel" aria-label="Run execution details">
    <div class="table-heading">
      <div>
        <h3>Run details</h3>
        <p>
          Execution context for {{ benchmark.description || benchmark.name }}. Missing fields were not
          recorded.
        </p>
      </div>
      <span class="count-badge">{{ profiles.length }} source{{ profiles.length === 1 ? '' : 's' }}</span>
    </div>
    <div class="details-overview">
      <div>
        <span>Suite commit</span><code>{{ report.metadata?.commit || 'Not recorded' }}</code>
      </div>
      <div>
        <span>Configured repetitions</span><strong>{{ benchmark.count ?? 'Not recorded' }}</strong>
      </div>
      <div>
        <span>Configured benchmark time</span><strong>{{ benchmark.benchtime || 'Not recorded' }}</strong>
      </div>
    </div>
    <details
      v-for="(profile, index) in profiles"
      :key="profile.id || index"
      class="source-details"
      :open="profiles.length === 1"
    >
      <summary>
        <ui-icon class="source-chevron" name="chevron" />
        <ui-icon name="cpu" /><span
          >{{ profile.samples[0]?.cpu || 'Runner CPU not recorded'
          }}<small>{{ profile.samples.length }} samples · {{ profile.name }}</small></span
        ><span class="count-badge">{{ profile.execution ? 'Execution captured' : 'Partial metadata' }}</span>
      </summary>
      <div class="source-body">
        <dl class="detail-grid">
          <div>
            <dt>Operating system / architecture</dt>
            <dd>{{ distinct(profile.samples, 'os') }} / {{ distinct(profile.samples, 'architecture') }}</dd>
          </div>
          <div>
            <dt>Go runtime (benchmark process)</dt>
            <dd>{{ profile.execution?.go_version || 'Not recorded' }}</dd>
          </div>
          <div>
            <dt>CPU count / GOMAXPROCS</dt>
            <dd>
              {{
                profile.execution
                  ? `${profile.execution.num_cpu} / ${profile.execution.gomaxprocs}`
                  : 'Not recorded'
              }}
            </dd>
          </div>
          <div>
            <dt>Execution captured at</dt>
            <dd>{{ profile.execution?.captured_at || 'Not recorded' }}</dd>
          </div>
        </dl>
        <template v-if="profile.execution">
          <details class="nested-details">
            <summary>Effective execution settings</summary>
            <dl class="key-values">
              <div v-for="(value, key) in profile.execution.settings" :key="key">
                <dt>{{ key }}</dt>
                <dd>{{ value === '' ? '(empty)' : value }}</dd>
              </div>
            </dl>
          </details>
          <details class="nested-details">
            <summary>Runner and container environment</summary>
            <dl class="key-values">
              <div v-for="(value, key) in profile.execution.environment" :key="key">
                <dt>{{ key }}</dt>
                <dd>{{ value }}</dd>
              </div>
            </dl>
          </details>
          <details class="nested-details">
            <summary>
              Binary inventory <span class="muted">{{ profile.execution.binaries?.length || 0 }} files</span>
            </summary>
            <p class="muted">
              Binaries available to this execution, including candidates that may not have been selected.
            </p>
            <div class="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Binary</th>
                    <th>Reference</th>
                    <th>Version / revision</th>
                    <th>SHA-256</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="binary in profile.execution.binaries"
                    :key="`${binary.project}/${binary.ref}/${binary.name}`"
                  >
                    <td>{{ binary.name }}</td>
                    <td>{{ binary.ref }}</td>
                    <td class="wrap-cell">{{ binary.version || 'Not recorded' }}</td>
                    <td class="wrap-cell">
                      <code>{{ binary.sha256 }}</code>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </details>
        </template>
        <p v-else class="notice">
          This source predates execution capture. Runtime settings, container identity and binary hashes are
          unavailable; the report exporter's environment is not substituted.
        </p>
      </div>
    </details>
    <details class="nested-details exporter-details">
      <summary>Report exporter (separate from execution)</summary>
      <dl class="key-values">
        <div v-for="(value, key) in report.generator" :key="key">
          <dt>{{ key }}</dt>
          <dd>{{ value }}</dd>
        </div>
      </dl>
    </details>
  </section>
</template>
<script>
import UiIcon from './UiIcon.vue';
import { sourceProfiles } from '../reports/benchmarks.mjs';
export default {
  components: { UiIcon },
  props: { report: { type: Object, required: true }, benchmark: { type: Object, required: true } },
  computed: {
    profiles() {
      return sourceProfiles(this.report, this.benchmark);
    },
  },
  methods: {
    distinct(samples, key) {
      return [...new Set(samples.map((s) => s[key] || 'Not recorded'))].join(', ');
    },
  },
};
</script>

<style scoped>
.details-overview {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 20px;
  padding: 0 24px 24px;
}
.details-overview span {
  display: block;
  font-size: 10px;
  color: var(--muted);
  margin-bottom: 9px;
}
.details-overview code {
  overflow-wrap: anywhere;
}
.details-overview strong {
  font-size: 12px;
  font-weight: 500;
}
.source-details {
  border-top: 1px solid var(--border);
}
.source-details > summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 24px;
  font-size: 12px;
}
.source-details > summary .count-badge {
  margin-left: auto;
}
.source-details small {
  display: block;
  color: var(--muted);
  font-size: 10px;
  margin-top: 5px;
  overflow-wrap: anywhere;
}
.source-body {
  padding: 0 24px 18px;
}
.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 24px;
  padding: 20px;
  background: var(--bg);
  border-radius: 6px;
}
.nested-details {
  margin-top: 12px;
}
.nested-details > summary {
  padding: 10px 0;
}
.nested-details p {
  font-size: 11px;
  margin: 10px 0;
}
.exporter-details {
  border-top: 1px solid var(--border);
  padding: 10px 24px;
  margin: 0;
}
@media (max-width: 560px) {
  .details-overview {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  .detail-grid {
    grid-template-columns: 1fr;
  }
  .source-details > summary {
    flex-wrap: wrap;
    padding: 18px;
  }
  .source-details > summary .count-badge {
    margin-left: 30px;
  }
  .source-body {
    padding: 0 18px 18px;
  }
}
.source-chevron {
  width: 12px;
}
.source-details[open] > summary .source-chevron {
  transform: rotate(90deg);
}
</style>
