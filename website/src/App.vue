<template>
  <div class="app-shell" :data-theme="theme">
    <router-view :theme="theme" @toggle-theme="toggleTheme" />
  </div>
</template>

<script>
import './styles/theme.css';
import './styles/shared.css';

export default {
  data() {
    let saved;
    try {
      saved = localStorage.getItem('benchmark-theme');
    } catch {
      /* Storage can be unavailable. */
    }
    return {
      theme: ['light', 'dark'].includes(saved)
        ? saved
        : window.matchMedia('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light',
    };
  },
  methods: {
    toggleTheme() {
      this.theme = this.theme === 'dark' ? 'light' : 'dark';
      try {
        localStorage.setItem('benchmark-theme', this.theme);
      } catch {
        /* Keep the in-memory preference. */
      }
    },
  },
};
</script>
