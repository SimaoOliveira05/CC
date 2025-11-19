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
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #00ff88;
  box-shadow: 0 0 10px rgba(0, 255, 136, 0.1);
  transition: all 0.3s;
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(0, 255, 136, 0.3);
  transform: translateY(-2px);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(0, 255, 136, 0.2);
}

.report-header h4 {
  color: #00ff88;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(0, 255, 136, 0.3);
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

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.metric-box {
  background: rgba(0, 255, 136, 0.05);
  border: 1px solid rgba(0, 255, 136, 0.2);
  border-radius: 6px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.metric-label {
  color: #a8b5c8;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.metric-value {
  color: #00ff88;
  font-weight: bold;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(0, 255, 136, 0.3);
}

@media (max-width: 500px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
