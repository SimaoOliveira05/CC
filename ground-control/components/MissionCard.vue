<template>
  <div class="mission-card" @click="$emit('select', mission)">
    <div class="mission-header">
      <h3>Missão #{{ mission.id }}</h3>
      <span class="mission-state" :class="sanitizeClass(mission.state)">
        {{ mission.state }}
      </span>
    </div>

    <div class="mission-info">
      <div class="info-row">
        <span class="label">Rover:</span>
        <span class="value">#{{ mission.idRover }}</span>
      </div>

      <div class="info-row">
        <span class="label">Tipo:</span>
        <span class="value">{{ getTaskTypeName(mission.taskType) }}</span>
      </div>

      <div class="info-row">
        <span class="label">Prioridade:</span>
        <span class="priority-badge" :class="'priority-' + mission.priority">
          {{ mission.priority }}
        </span>
      </div>

      <div class="info-row">
        <span class="label">Coordenadas:</span>
        <span class="value coordinate">{{ formatCoordinate(mission.coordinate) }}</span>
      </div>

      <div class="info-row">
        <span class="label">Reports:</span>
        <span class="value">{{ mission.reports.length }}</span>
      </div>

      <div class="info-row">
        <span class="label">Última atualização:</span>
        <span class="value">{{ formatTime(mission.lastUpdate) }}</span>
      </div>
    </div>

    <div class="mission-footer">
      <span class="click-hint">Clique para ver detalhes</span>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  mission: Object,
  required: true
});

defineEmits(['select']);

const taskTypes = {
  0: 'Captura de Imagem',
  1: 'Coleta de Amostra',
  2: 'Análise Ambiental',
  3: 'Reparação/Resgate',
  4: 'Mapeamento Topográfico',
  5: 'Instalação'
};

const getTaskTypeName = (type) => {
  return taskTypes[type] || 'Desconhecido';
};

const formatTime = (dateString) => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleTimeString('pt-PT', { hour: '2-digit', minute: '2-digit' });
};

const formatCoordinate = (coord) => {
  if (!coord || coord.latitude === undefined) return 'N/A';
  return `(${coord.latitude.toFixed(4)}, ${coord.longitude.toFixed(4)})`;
};

const sanitizeClass = (state) => {
  // Converter para formato válido de classe CSS
  if (!state) return '';
  // Remover espaços, converter para minúsculas mas manter compatibilidade
  return state.toLowerCase().replace(/\s+/g, '-');
};
</script>

<style scoped>
.mission-card {
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border: 2px solid #1e5a96;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 0 10px rgba(30, 90, 150, 0.2);
}

.mission-card:hover {
  border-color: #00d4ff;
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.4);
  transform: translateY(-5px);
}

.mission-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 15px;
  gap: 10px;
}

.mission-header h3 {
  color: #00d4ff;
  font-size: 18px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
  margin: 0;
}

.mission-state {
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
  white-space: nowrap;
}

/* Pending states */
.mission-state.pending,
.mission-state.Pending {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
  border: 1px solid #ffaa00;
}

/* Moving to states */
.mission-state.moving-to,
.mission-state.moving\ to,
.mission-state.Moving\ to {
  background: rgba(255, 170, 0, 0.2);
  color: #ff9500;
  border: 1px solid #ff9500;
}

/* In Progress states */
.mission-state.in\ progress,
.mission-state.In\ Progress,
.mission-state.inprogress,
.mission-state.InProgress,
.mission-state.in-progress {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
  border: 1px solid #ff4444;
}

/* Completed states */
.mission-state.completed,
.mission-state.Completed {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  border: 1px solid #00ff88;
}

.mission-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 15px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
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

.priority-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
}

.priority-1 {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
  border: 1px solid #ff4444;
}

.priority-2 {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
  border: 1px solid #ffaa00;
}

.priority-3 {
  background: rgba(0, 212, 255, 0.2);
  color: #00d4ff;
  border: 1px solid #00d4ff;
}

.mission-footer {
  text-align: center;
  padding-top: 15px;
  border-top: 1px solid rgba(0, 212, 255, 0.2);
}

.click-hint {
  font-size: 12px;
  color: #a8b5c8;
  font-style: italic;
}
</style>
