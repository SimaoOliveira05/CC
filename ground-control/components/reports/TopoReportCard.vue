<template>
  <div class="report-card topo-report">
    <div class="report-header">
      <h4>🗺️ Mapeamento Topográfico</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="coordinates">
        <div class="coord-item">
          <span class="label">Latitude</span>
          <span class="value">{{ report.latitude.toFixed(6) }}°</span>
        </div>
        <div class="coord-item">
          <span class="label">Longitude</span>
          <span class="value">{{ report.longitude.toFixed(6) }}°</span>
        </div>
      </div>
      <div class="height-info">
        <span class="label">Altura</span>
        <span class="value">{{ report.height.toFixed(2) }} m</span>
      </div>
      <div class="map-link">
        <a :href="`https://maps.google.com/?q=${report.latitude},${report.longitude}`" 
           target="_blank" 
           class="btn-map">
          Ver no Mapa 🌐
        </a>
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
  border-left: 3px solid var(--accent-warning);
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
  color: var(--accent-warning);
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

.coordinates {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

.coord-item {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.height-info {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  color: var(--text-secondary);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.value {
  color: var(--accent-warning);
  font-weight: 500;
  font-size: 13px;
}

.map-link {
  margin-top: 6px;
}

.btn-map {
  display: inline-block;
  background: var(--accent-warning);
  color: white;
  padding: 6px 10px;
  border-radius: var(--radius-sm);
  text-decoration: none;
  font-weight: 500;
  font-size: 11px;
  text-align: center;
  transition: opacity 0.2s;
}

.btn-map:hover {
  opacity: 0.9;
}
</style>
