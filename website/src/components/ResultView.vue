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
    async updateMetadata() {
      this.metadataLinks = [];
      const resultName = this.$route.params.result;

      try {
        const ghaEvent = await import(`../../public/result/${resultName}/gha-event.json`);
        const env = await import(`../../public/result/${resultName}/env.txt`);

        let repo = ghaEvent.repository?.full_name;
        let commit = ghaEvent.commits?.[ghaEvent.commits.length - 1]?.id;
        let runId, runAttempt;

        const envs = ((envString) => {
          return envString.split('\n').filter(line => line).map(line => {
            const [key, value] = line.split('=');
            return { key, value };
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

        if (commit) {
          this.metadataLinks.push({
            text: `Commit ${commit.substring(0, 7)}`,
            url: `https://github.com/${repo}/commit/${commit}`,
            icon: require('../assets/github.svg')
          });
        }
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
