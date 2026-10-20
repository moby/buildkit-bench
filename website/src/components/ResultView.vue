<template>
  <div class="result">
    <header class="result-header" v-if="metadataLinks.length > 0">
      <div class="metadata-links">
        <a v-for="link in metadataLinks" :key="link.url" :href="link.url" target="_blank" class="metadata-link">
          <img :src="link.icon" class="icon" /> {{ link.text }}
        </a>
      </div>
    </header>
    <nav v-if="projects.length > 1" class="tabs" aria-label="Benchmark project">
      <button v-for="project in projects" :key="project" :class="{ active: selectedProject === project }"
        :aria-pressed="selectedProject === project" @click="activeTab = project">
        {{ project === 'buildkit' ? 'BuildKit' : 'Buildx' }}
      </button>
    </nav>
    <iframe v-if="!reportUrl && originalUrl" class="legacy-report" :src="originalUrl"
      :title="`${selectedProject} benchmark report`"></iframe>
    <div v-else class="report-content" :aria-busy="loading">
      <p v-if="loading" role="status">Loading report...</p>
      <div v-else-if="error" role="alert">
        <p>{{ error }}</p>
        <button @click="loadReport">Retry</button>
      </div>
      <p v-else-if="!reportUrl">This run has no report available.</p>
      <p v-else-if="charts.length === 0">This report has no measurements.</p>
      <benchmark-chart v-for="chart in charts" :key="`${reportUrl}/${chart.id}`" :option="chart.option" />
    </div>
  </div>
</template>

<script>
import resultsData from '../assets/results.json';
import { markRaw } from 'vue';
import BenchmarkChart from './BenchmarkChart.vue';
import { reportCharts } from '../charts/report.mjs';

export default {
  name: 'ResultView',
  components: { BenchmarkChart },
  data() {
    return {
      metadataLinks: [],
      activeTab: 'buildkit',
      charts: [],
      loading: false,
      error: '',
      request: null,
    };
  },
  computed: {
    projects() {
      const result = this.$route.params.result;
      return ['buildkit', 'buildx'].filter(project =>
        resultsData.jsonReports[result]?.includes(project) || resultsData.htmlReports[result]?.[project]);
    },
    selectedProject() {
      return this.projects.includes(this.activeTab) ? this.activeTab : this.projects[0];
    },
    reportUrl() {
      const result = this.$route.params.result;
      return resultsData.jsonReports[result]?.includes(this.selectedProject)
        ? `${process.env.BASE_URL}result/${result}/${this.selectedProject}.json` : '';
    },
    originalUrl() {
      const result = this.$route.params.result;
      const file = resultsData.htmlReports[result]?.[this.selectedProject];
      return file ? `${process.env.BASE_URL}result/${result}/${file}` : '';
    }
  },
  created() {
    this.updateMetadata();
  },
  beforeUnmount() {
    this.request?.abort();
  },
  watch: {
    '$route.params.result': function() {
      this.updateMetadata();
    },
    reportUrl: { immediate: true, handler: 'loadReport' },
  },
  methods: {
    async loadReport() {
      this.request?.abort();
      const request = markRaw(new AbortController());
      this.request = request;
      this.charts = [];
      this.error = '';
      this.loading = !!this.reportUrl;
      if (!this.reportUrl) return;
      const project = this.selectedProject;
      try {
        const response = await fetch(this.reportUrl, { signal: request.signal });
        if (!response.ok) throw new Error(`Report request failed (${response.status})`);
        const report = await response.json();
        if (request.signal.aborted) return;
        if (report.project !== project) throw new Error('Report project does not match the selection');
        this.charts = markRaw(reportCharts(report));
      } catch (error) {
        if (!request.signal.aborted) this.error = `Unable to load report: ${error.message}`;
      } finally {
        if (!request.signal.aborted) this.loading = false;
      }
    },
    async updateMetadata() {
      this.metadataLinks = [];
      const resultName = this.$route.params.result;

      try {
        const ghaEvent = await import(`../../public/result/${resultName}/gha-event.json`);
        const env = await import(`../../public/result/${resultName}/env.txt`);
        if (resultName !== this.$route.params.result) return;

        let repo = ghaEvent.repository?.full_name;
        let commit = ghaEvent.commits?.[ghaEvent.commits.length - 1]?.id;
        let runId, runAttempt;

        const envs = ((envString) => {
          return envString.split('\n').filter(line => line).map(line => {
            const [key, value] = line.split('=');
            return {key, value};
          });
        })(env.default);

        const repoEnv = envs.find(entry => entry.key === 'GITHUB_REPOSITORY');
        const commitEnv = envs.find(entry => entry.key === 'GITHUB_SHA');
        const runIdEnv = envs.find(entry => entry.key === 'GITHUB_RUN_ID');
        const runAttemptEnv = envs.find(entry => entry.key === 'GITHUB_RUN_ATTEMPT');
        if (!repo && repoEnv) repo = repoEnv.value;
        if (!commit && commitEnv) commit = commitEnv.value;
        if (runIdEnv) runId = runIdEnv.value;
        if (runAttemptEnv) runAttempt = runAttemptEnv.value;
        if (!repo) return;

        const links = [];
        if (commit) {
          links.push({
            text: `Commit ${commit.substring(0, 7)}`,
            url: `https://github.com/${repo}/commit/${commit}`,
            icon: require('../assets/github.svg')
          });
        }
        if (runId && runAttempt) {
          links.push({
            text: `GitHub Actions Run`,
            url: `https://github.com/${repo}/actions/runs/${runId}/attempts/${runAttempt}`,
            icon: require('../assets/github.svg')
          });
        }
        links.push({
          text: `Logs`,
          url: `${process.env.BASE_URL}result/${resultName}/logs.tar.gz`,
          icon: require('../assets/logs.svg')
        });
        this.metadataLinks = links;
      } catch (error) {
        console.error(`failed to load metadata for ${resultName}`, error);
      }
    }
  }
};
</script>

<style scoped>
.result {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.result-header {
  flex: 0 0 auto;
  background-color: #fff;
  padding: 10px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.metadata-links {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
}

.metadata-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: #007bff;
  font-weight: bold;
  transition: color 0.3s, transform 0.3s;
}

.metadata-link:hover {
  color: #0056b3;
  transform: translateY(-3px);
}

.icon {
  width: 20px;
  height: 20px;
  margin-right: 8px;
}

.report-content {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 16px;
}

.legacy-report {
  flex: 1 1 auto;
  min-height: 0;
  width: 100%;
  border: 0;
}

.tabs {
  display: flex;
  list-style: none;
  padding: 0;
  margin: 0;
  background-color: #f8f9fa;
  border-bottom: 1px solid #dee2e6;
}

.tabs button {
  border: 0;
  background: transparent;
  font: inherit;
  padding: 10px 20px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.tabs button.active {
  background-color: #e9ecef;
  font-weight: bold;
}

.tabs button:hover {
  background-color: #e9ecef;
}

</style>
