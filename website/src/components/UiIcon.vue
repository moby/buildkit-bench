<template>
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="paths[name] || paths.chart" />
  </svg>
</template>
<script>
const paths = {
  chart: 'M4 20V10m5 10V4m6 16v-7m5 7V7M2 22h20',
  history: 'M3 11a9 9 0 1 1 2 7M3 4v7h7m2-5v6l4 2',
  search: 'M21 21l-5-5M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0',
  sun: 'M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1 1m12 12 1 1M5 19l1-1M18 6l1-1M17 12a5 5 0 1 1-10 0 5 5 0 0 1 10 0',
  moon: 'M21 13A9 9 0 0 1 11 3a9 9 0 1 0 10 10',
  external: 'M14 3h7v7m0-7L10 14m-1-9H3v16h16v-6',
  info: 'M12 11v6m0-10h.01M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0',
  cpu: 'M6 6h12v12H6zM9 9h6v6H9zM9 2v4m6-4v4M9 18v4m6-4v4M2 9h4m-4 6h4m12-6h4m-4 6h4',
  git: 'M6 6v12m0-12a2 2 0 1 0 0-4 2 2 0 0 0 0 4m0 16a2 2 0 1 0 0-4 2 2 0 0 0 0 4m12-16a2 2 0 1 0 0-4 2 2 0 0 0 0 4m0 0v3c0 5-12 1-12 7',
  arrow: 'M5 12h14m-5-5 5 5-5 5',
  chevron: 'm8 5 7 7-7 7',
};
export default { props: { name: String }, data: () => ({ paths }) };
</script>
