<template>
  <div class="report-card env-report">
    <div class="report-header">
      <h4>🌍 Análise Ambiental</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="metrics-grid">
        <div class="metric-box">
          <span class="metric-label">Temperatura</span>
          <span class="metric-value">{{ report.temp.toFixed(2) }}°C</span>
        </div>
        <div class="metric-box">
          <span class="metric-label">Oxigênio</span>
          <span class="metric-value">{{ report.oxygen.toFixed(2) }}%</span>
        </div>
        <div class="metric-box">
          <span class="metric-label">Pressão</span>
          <span class="metric-value">{{ report.pressure.toFixed(2) }} hPa</span>
        </div>
        <div class="metric-box">
          <span class="metric-label">Humidade</span>
          <span class="metric-value">{{ report.humidity.toFixed(2) }}%</span>
        </div>
        <div class="metric-box">
          <span class="metric-label">Velocidade do Vento</span>
          <span class="metric-value">{{ report.windSpeed.toFixed(2) }} m/s</span>
        </div>
        <div class="metric-box">
          <span class="metric-label">Radiação</span>
          <span class="metric-value">{{ report.radiation.toFixed(2) }} µSv</span>
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
  border-left: 3px solid var(--accent-success);
  transition: background 0.2s;
}

.report-card:hover {
  background: var(--bg-hover);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.report-header h4 {
  color: var(--accent-success);
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

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.metric-box {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-label {
  color: var(--text-secondary);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.metric-value {
  color: var(--accent-success);
  font-weight: 500;
  font-size: 13px;
}

@media (max-width: 500px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
