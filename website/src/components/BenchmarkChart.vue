<template>
  <v-chart class="benchmark-chart" :option="option" :update-options="{ notMerge: true }" autoresize />
</template>

<script>
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { BarChart, BoxplotChart } from 'echarts/charts';
import { DataZoomComponent, GridComponent, LegendComponent, MarkLineComponent, TitleComponent, ToolboxComponent, TooltipComponent } from 'echarts/components';
import VChart from 'vue-echarts';

use([CanvasRenderer, BarChart, BoxplotChart, DataZoomComponent, GridComponent,
  LegendComponent, MarkLineComponent, TitleComponent, ToolboxComponent, TooltipComponent]);

export default {
  components: { VChart },
  props: { option: { type: Object, required: true } },
};
</script>

<style scoped>
.benchmark-chart {
  width: 100%;
  height: 500px;
}
</style>
