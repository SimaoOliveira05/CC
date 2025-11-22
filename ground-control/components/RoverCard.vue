<template>
  <div class="rover-card">
    <div class="rover-header">
      <span class="rover-id">ROVER #{{ rover.id }}</span>
      <span class="status-badge" :class="rover.state.toLowerCase()">
        {{ rover.state }}
      </span>
    </div>
    
    <div class="rover-body">
      <div class="metric">
        <span class="label">Bateria</span>
        <div class="battery-bar">
          <div class="battery-fill" :style="{ width: rover.battery + '%', background: getBatteryColor() }"></div>
        </div>
        <span class="value">{{ rover.battery }}%</span>
      </div>

      <div class="metric">
        <span class="label">Velocidade</span>
        <span class="value">{{ rover.speed.toFixed(2) }} m/s</span>
      </div>

      <div class="metric">
        <span class="label">Posição</span>
        <span class="value coordinate">{{ formatPosition(rover.position) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  rover: Object,
  required: true
});

const getBatteryColor = () => {
  if (props.rover.battery > 60) return '#00ff88';
  if (props.rover.battery > 30) return '#ffaa00';
  return '#ff4444';
};

const formatPosition = (pos) => {
  if (!pos || pos.latitude === undefined || pos.longitude === undefined) return 'N/A';
  return `(${pos.latitude.toFixed(4)}, ${pos.longitude.toFixed(4)})`;
};
</script>

<style scoped>
.rover-card {
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border: 2px solid #00d4ff;
  border-radius: 8px;
  padding: 15px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.2);
}

.rover-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.5);
  border-color: #00ff88;
}

.rover-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.rover-id {
  font-weight: bold;
  color: #00d4ff;
  font-size: 13px;
  text-shadow: 0 0 5px rgba(0, 212, 255, 0.5);
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
}

.status-badge.active {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  border: 1px solid #00ff88;
}

.status-badge.inactive {
  background: rgba(100, 100, 100, 0.2);
  color: #aaa;
  border: 1px solid #666;
}

.status-badge.inmission {
  background: rgba(255, 107, 31, 0.2);
  color: #ff6b1f;
  border: 1px solid #ff6b1f;
}

.rover-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.label {
  font-size: 11px;
  color: #a8b5c8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.value {
  font-size: 14px;
  color: #e8eef7;
  font-weight: bold;
}

.battery-bar {
  height: 8px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid rgba(0, 212, 255, 0.3);
}

.battery-fill {
  height: 100%;
  transition: width 0.3s, background 0.3s;
  border-radius: 4px;
}

.coordinate {
  color: #00d4ff;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}
</style>
