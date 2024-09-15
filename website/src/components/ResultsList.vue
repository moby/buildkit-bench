<template>
  <div class="sidebar">
    <h2>Results</h2>
    <ul>
      <li v-for="result in results" :key="result" @click="loadResult(result)"
          :class="['result-item', { 'selected': result === selectedResult }]">
        {{ result }}
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  data() {
    return {
      results: [],
      selectedResult: null
    };
  },
  created() {
    fetch('results.json')
        .then(response => response.json())
        .then(data => {
          this.results = data.results.sort((a, b) => b.localeCompare(a));
          if (this.results.length > 0) {
            this.selectedResult = this.results[0];
            this.$router.push(`/result/${this.selectedResult}`).catch(err => {
              if (err.name !== 'NavigationDuplicated') {
                throw err;
              }
            });
          }
        });
    this.updateSelectedResult();
  },
  watch: {
    '$route.params.result': 'updateSelectedResult'
  },
  methods: {
    loadResult(result) {
      const targetRoute = `/result/${result}`;
      if (this.$route.path !== targetRoute) {
        this.selectedResult = result;
        this.$router.push(targetRoute).catch(err => {
          if (err.name !== 'NavigationDuplicated') {
            throw err;
          }
        });
      }
    },
    updateSelectedResult() {
      if (this.$route.params.result) {
        this.selectedResult = this.$route.params.result;
      } else if (this.results.length > 0) {
        this.selectedResult = this.results[0];
      }
    },
  }
};
</script>

<style scoped>
.sidebar {
  width: 250px;
  background-color: #f8f9fa;
  padding: 15px;
  box-shadow: 2px 0 5px rgba(0, 0, 0, 0.1);
  height: 100vh;
  position: fixed;
}

.sidebar h2 {
  font-size: 1.5em;
  margin-bottom: 20px;
  color: #343a40;
}

.sidebar ul {
  list-style-type: none;
  padding: 0;
}

.sidebar .result-item {
  padding: 10px;
  margin-bottom: 10px;
  background-color: #ffffff;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.3s;
}

.sidebar .result-item:hover {
  background-color: #e9ecef;
  transform: translateX(5px);
}

.sidebar .result-item:active {
  background-color: #dee2e6;
}

.sidebar .result-item.selected {
  background-color: #d4edda;
  border-left: 5px solid #28a745;
}
</style>
