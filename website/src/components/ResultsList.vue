<template>
  <div class="results">
    <h2>Results</h2>
    <ul>
      <li v-for="result in results" :key="result" @click="loadResult(result)" :class="['result-item', { 'selected': result === selectedResult }]">
        <i class="calendar-icon"></i>
        {{ formatResult(result) }}
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
    formatResult(result) {
      const year = result.substring(0, 4);
      const month = result.substring(4, 6);
      const day = result.substring(6, 8);
      const hour = result.substring(9, 11);
      const minute = result.substring(11, 13);
      const second = result.substring(13, 15);
      return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
    }
  }
};
</script>

<style scoped>
.results h2 {
  font-size: 1.5em;
  margin-bottom: 20px;
  color: #343a40;
}

.results ul {
  list-style-type: none;
  padding: 0;
}

.results .result-item {
  display: flex;
  align-items: center;
  padding: 10px;
  margin-bottom: 10px;
  background-color: #ffffff;
  border-radius: 5px;
  cursor: pointer;
  transition: background-color 0.3s, transform 0.3s;
}

.results .result-item:hover {
  background-color: #e9ecef;
  transform: translateX(5px);
}

.results .result-item:active {
  background-color: #dee2e6;
}

.results .result-item.selected {
  background-color: #d4edda;
  border-left: 5px solid #28a745;
}

.results .calendar-icon {
  margin-right: 10px;
  width: 16px;
  height: 16px;
  background-image: url('../assets/calendar.svg');
  background-size: contain;
  background-repeat: no-repeat;
}
</style>
