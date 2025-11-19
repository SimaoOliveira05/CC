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
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #ff9800;
  box-shadow: 0 0 10px rgba(255, 152, 0, 0.1);
  transition: all 0.3s;
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(255, 152, 0, 0.3);
  transform: translateY(-2px);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 152, 0, 0.2);
}

.report-header h4 {
  color: #ffa726;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(255, 167, 38, 0.3);
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

.coordinates {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.coord-item {
  background: rgba(255, 152, 0, 0.05);
  border: 1px solid rgba(255, 152, 0, 0.2);
  border-radius: 6px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.height-info {
  background: rgba(255, 152, 0, 0.05);
  border: 1px solid rgba(255, 152, 0, 0.2);
  border-radius: 6px;
  padding: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.label {
  color: #a8b5c8;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.value {
  color: #ffa726;
  font-weight: bold;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(255, 167, 38, 0.3);
}

.map-link {
  margin-top: 8px;
}

.btn-map {
  display: inline-block;
  background: linear-gradient(135deg, #ff9800, #ffa726);
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  text-decoration: none;
  font-weight: bold;
  font-size: 12px;
  text-align: center;
  transition: all 0.3s;
  border: none;
  cursor: pointer;
}

.btn-map:hover {
  background: linear-gradient(135deg, #fb8c00, #ffb74d);
  box-shadow: 0 0 10px rgba(255, 152, 0, 0.5);
}
</style>
