<template>
  <div class="report-card sample-report">
    <div class="report-header">
      <h4>🧪 Coleta de Amostra</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="info-item">
        <span class="label">Componentes:</span>
        <span class="value">{{ report.numSamples }}</span>
      </div>
      <div class="components-list">
        <div v-for="(comp, idx) in report.components" :key="idx" class="component-item">
          <span class="comp-name">{{ comp.name }}</span>
          <div class="comp-bar">
            <div class="comp-fill" :style="{ width: comp.percentage + '%' }"></div>
          </div>
          <span class="comp-value">{{ comp.percentage.toFixed(2) }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  report: Object,
  required: true
});
</script>

<style scoped>
.report-card {
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #9c27b0;
  box-shadow: 0 0 10px rgba(156, 39, 176, 0.1);
  transition: all 0.3s;
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(156, 39, 176, 0.3);
  transform: translateY(-2px);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(156, 39, 176, 0.2);
}

.report-header h4 {
  color: #bb86fc;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(187, 134, 252, 0.3);
}

.last-badge {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
}

.report-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  margin-bottom: 8px;
}

.label {
  color: #a8b5c8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.value {
  color: #e8eef7;
  font-weight: bold;
}

.components-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.component-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.comp-name {
  color: #bb86fc;
  font-weight: bold;
  min-width: 40px;
}

.comp-bar {
  flex: 1;
  height: 6px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 3px;
  overflow: hidden;
  border: 1px solid rgba(187, 134, 252, 0.3);
}

.comp-fill {
  height: 100%;
  background: linear-gradient(90deg, #bb86fc, #9c27b0);
  border-radius: 3px;
}

.comp-value {
  color: #bb86fc;
  font-weight: bold;
  min-width: 45px;
  text-align: right;
}
</style>
