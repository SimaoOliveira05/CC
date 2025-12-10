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
  background: var(--bg-primary);
  border-radius: var(--radius-md);
  padding: 14px;
  border-left: 3px solid #a855f7;
  transition: background 0.2s;
}

.report-card:hover {
  background: var(--bg-hover);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.report-header h4 {
  color: #a855f7;
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.last-badge {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 500;
  text-transform: uppercase;
}

.report-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  margin-bottom: 6px;
}

.label {
  color: var(--text-secondary);
  text-transform: uppercase;
  font-size: 10px;
  letter-spacing: 0.3px;
}

.value {
  color: var(--text-primary);
  font-weight: 500;
}

.components-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.component-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.comp-name {
  color: #a855f7;
  font-weight: 500;
  min-width: 36px;
}

.comp-bar {
  flex: 1;
  height: 5px;
  background: var(--bg-hover);
  border-radius: 3px;
  overflow: hidden;
}

.comp-fill {
  height: 100%;
  background: linear-gradient(90deg, #a855f7, #9333ea);
  border-radius: 3px;
}

.comp-value {
  color: #a855f7;
  font-weight: 500;
  min-width: 42px;
  text-align: right;
}
</style>
