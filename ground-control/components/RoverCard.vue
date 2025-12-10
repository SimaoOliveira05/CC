<template>
  <div class="rover-card" :title="`Rover #${rover.id} - ${formatPosition(rover.position)}`">
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

      <div class="metric">
        <span class="label">Missões na Fila</span>
        <div class="queue-info">
          <span class="queue-badge priority-1">
            P1: {{ getQueueCount(1) }}
          </span>
          <span class="queue-badge priority-2">
            P2: {{ getQueueCount(2) }}
          </span>
          <span class="queue-badge priority-3">
            P3: {{ getQueueCount(3) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { watch } from 'vue';

const props = defineProps({
  rover: Object,
  required: true
});

// Watch for changes in rover data
watch(() => props.rover, (newVal) => {
  console.log('Rover updated:', newVal.id, 'queuedMissions:', newVal.queuedMissions);
}, { deep: true });

const getBatteryColor = () => {
  if (props.rover.battery > 60) return '#00ff88';
  if (props.rover.battery > 30) return '#ffaa00';
  return '#ff4444';
};

const formatPosition = (pos) => {
  if (!pos || pos.latitude === undefined || pos.longitude === undefined) return 'N/A';
  return `(${pos.latitude.toFixed(4)}, ${pos.longitude.toFixed(4)})`;
};

const getQueueCount = (priority) => {
  if (!props.rover.queuedMissions) return 0;
  if (priority === 1) return props.rover.queuedMissions.priority1Count || 0;
  if (priority === 2) return props.rover.queuedMissions.priority2Count || 0;
  if (priority === 3) return props.rover.queuedMissions.priority3Count || 0;
  return 0;
};
</script>

<style scoped>
.rover-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 14px;
  contain: content;
}

.rover-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rover-id {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
}

.status-badge {
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
}

.status-badge.active {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}

.status-badge.inactive {
  background: var(--bg-hover);
  color: var(--text-muted);
}

.status-badge.inmission {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.rover-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.value {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
  font-variant-numeric: tabular-nums;
}

.queue-info {
  display: flex;
  gap: 6px;
}

.queue-badge {
  padding: 3px 6px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 500;
  font-variant-numeric: tabular-nums;
  min-width: 42px;
  text-align: center;
}

.queue-badge.priority-1 {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}

.queue-badge.priority-2 {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.queue-badge.priority-3 {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-primary);
}

.battery-bar {
  height: 6px;
  background: var(--bg-hover);
  border-radius: 3px;
  overflow: hidden;
}

.battery-fill {
  height: 100%;
  border-radius: 3px;
}

.coordinate {
  color: var(--text-secondary);
  font-family: monospace;
  font-size: 12px;
  font-variant-numeric: tabular-nums;
}
</style>
