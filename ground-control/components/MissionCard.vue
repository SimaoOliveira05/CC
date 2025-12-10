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
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  cursor: pointer;
  transition: border-color 0.2s;
}

.mission-card:hover {
  border-color: var(--accent-primary);
}

.mission-header {
  display: flex;
  justify-content: space-between;
  align-items: start;
  margin-bottom: 14px;
  gap: 10px;
}

.mission-header h3 {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.mission-state {
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
  white-space: nowrap;
}

.mission-state.pending,
.mission-state.Pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.mission-state.moving-to,
.mission-state.moving\ to,
.mission-state.Moving\ to {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.mission-state.in\ progress,
.mission-state.In\ Progress,
.mission-state.inprogress,
.mission-state.InProgress,
.mission-state.in-progress {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-primary);
}

.mission-state.completed,
.mission-state.Completed {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}

.mission-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 14px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.label {
  color: var(--text-secondary);
  text-transform: uppercase;
  font-size: 11px;
  letter-spacing: 0.3px;
}

.value {
  color: var(--text-primary);
  font-weight: 500;
}

.priority-badge {
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
}

.priority-1 {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}

.priority-2 {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.priority-3 {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-primary);
}

.mission-footer {
  text-align: center;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.click-hint {
  font-size: 12px;
  color: var(--text-muted);
}
</style>
