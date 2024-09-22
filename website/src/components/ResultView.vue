<template>
  <div class="result">
    <header class="result-header" v-if="metadataLinks.length > 0">
      <div class="metadata-links">
        <a v-for="link in metadataLinks" :key="link.url" :href="link.url" target="_blank" class="metadata-link">
          <img :src="link.icon" class="icon" /> {{ link.text }}
        </a>
      </div>
    </header>
    <div class="iframe-container">
      <iframe :src="resultUrl"></iframe>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ResultView',
  data() {
    return {
      metadataLinks: []
    };
  },
  computed: {
    resultUrl() {
      return `${process.env.BASE_URL}result/${this.$route.params.result}/index.html`;
    }
  },
  created() {
    this.updateMetadata();
  },
  watch: {
    '$route.params.result': 'updateMetadata'
  },
  methods: {
    updateMetadata() {
      this.metadataLinks = [];
      const resultName = this.$route.params.result;

      const ghaEvent = fetch(`${process.env.BASE_URL}result/${resultName}/gha-event.json`)
        .then(response => response.ok ? response.json() : null);
      const env = fetch(`${process.env.BASE_URL}result/${resultName}/env.txt`)
        .then(response => response.ok ? response.text() : null);

      Promise.all([ghaEvent, env])
        .then(([ghaEvent, env]) => {
          let repo = undefined;
          let commit = undefined;
          let runId = undefined;
          let runAttempt = undefined;
          if (ghaEvent) {
            repo = ghaEvent.repository?.full_name;
            if (ghaEvent.commits && ghaEvent.commits.length > 0) {
              commit = ghaEvent.commits[ghaEvent.commits.length - 1].id;
            }
          }
          if (env) {
            const repoMatch = env.match(/GITHUB_REPOSITORY=(.+)/);
            const commitMatch = env.match(/GITHUB_SHA=(.+)/);
            const runIdMatch = env.match(/GITHUB_RUN_ID=(\d+)/);
            const runAttemptMatch = env.match(/GITHUB_RUN_ATTEMPT=(\d+)/);
            if (!repo && repoMatch) {
              repo = repoMatch[1];
            }
            if (!commit && commitMatch) {
              commit = commitMatch[1];
            }
            if (runIdMatch) {
              runId = runIdMatch[1];
            }
            if (runAttemptMatch) {
              runAttempt = runAttemptMatch[1];
            }
          }
          if (!repo || !commit) {
            return;
          }
          this.metadataLinks.push({
            text: `Commit ${commit.substring(0, 7)}`,
            url: `https://github.com/${repo}/commit/${commit}`,
            icon: require('../assets/github.svg')
          });
          if (runId && runAttempt) {
            this.metadataLinks.push({
              text: `GitHub Actions Run`,
              url: `https://github.com/${repo}/actions/runs/${runId}/attempts/${runAttempt}`,
              icon: require('../assets/github.svg')
            });
          }
          this.metadataLinks.push({
            text: `Logs`,
            url: `https://github.com/${repo}/tree/gh-pages/result/${resultName}/logs`,
            icon: require('../assets/logs.svg')
          });
        })
        .catch(() => {});
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

.iframe-container {
  flex: 1 1 auto;
  overflow: hidden;
}

iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
