<template>
  <div class="mission-detail">
    <!-- Mission Header -->
    <div class="detail-header">
      <div class="mission-title">
        <h2>Missão #{{ mission.id }}</h2>
        <span class="state-badge" :class="sanitizeClass(mission.state)">{{ mission.state }}</span>
      </div>

      <div class="mission-meta">
        <div class="meta-item">
          <span class="meta-label">Rover Atribuído</span>
          <span class="meta-value">#{{ mission.idRover }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Tipo de Tarefa</span>
          <span class="meta-value">{{ getTaskTypeName(mission.taskType) }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Prioridade</span>
          <span class="meta-value priority" :class="'p-' + mission.priority">{{ mission.priority }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Coordenadas Destino</span>
          <span class="meta-value coordinate">{{ formatCoordinate(mission.coordinate) }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Criada em</span>
          <span class="meta-value">{{ formatDate(mission.createdAt) }}</span>
        </div>
        <div class="meta-item">
          <span class="meta-label">Última Atualização</span>
          <span class="meta-value">{{ formatDate(mission.lastUpdate) }}</span>
        </div>
      </div>
    </div>

    <!-- Assembled Image Section (for image capture missions) -->
    <div v-if="mission.taskType === 0 && mission.assembledImage" class="assembled-image-section">
      <h3>🖼️ Imagem Reconstruída</h3>
      <div class="assembled-image-container">
        <img :src="`data:image/jpeg;base64,${mission.assembledImage}`" alt="Imagem Reconstruída" />
        <p class="image-info">{{ mission.reports.length }} chunks recebidos | {{ formatImageSize(mission.assembledImage) }}</p>
      </div>
    </div>
    
    <!-- Debug info -->
    <div v-if="mission.taskType === 0 && !mission.assembledImage && mission.reports.length > 0" class="debug-info">
      ⚠️ Imagem ainda não foi reassembled ({{ mission.reports.length }} chunks recebidos)
    </div>

    <!-- Reports Section -->
    <div class="reports-section">
      <h3>📦 Reports Individuais ({{ mission.reports.length }})</h3>

      <div v-if="mission.reports.length === 0" class="no-reports">
        <p>Nenhum report recebido ainda</p>
      </div>

      <div v-else class="reports-list">
        <component 
          v-for="(report, index) in mission.reports" 
          :key="index"
          :is="getReportComponent(report)"
          :report="report"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import ImageReportCard from './reports/ImageReportCard.vue';
import SampleReportCard from './reports/SampleReportCard.vue';
import EnvReportCard from './reports/EnvReportCard.vue';
import RepairReportCard from './reports/RepairReportCard.vue';
import TopoReportCard from './reports/TopoReportCard.vue';
import InstallReportCard from './reports/InstallReportCard.vue';

const props = defineProps({
  mission: Object,
  required: true
});

const taskTypes = {
  0: 'Captura de Imagem',
  1: 'Coleta de Amostra',
  2: 'Análise Ambiental',
  3: 'Reparação/Resgate',
  4: 'Mapeamento Topográfico',
  5: 'Instalação'
};

const getTaskTypeName = (type) => taskTypes[type] || 'Desconhecido';

const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleString('pt-PT');
};

const formatCoordinate = (coord) => {
  if (!coord || coord.latitude === undefined || coord.longitude === undefined) return 'N/A';
  return `(${coord.latitude.toFixed(4)}, ${coord.longitude.toFixed(4)})`;
};

const sanitizeClass = (state) => {
  // Converter para formato válido de classe CSS
  if (!state) return '';
  // Converter para minúsculas e substituir espaços por hífen
  return state.toLowerCase().replace(/\s+/g, '-');
};

const formatImageSize = (base64String) => {
  if (!base64String) return '0 KB';
  // Base64 string length * 0.75 gives approximate byte size
  const bytes = base64String.length * 0.75;
  const kb = bytes / 1024;
  if (kb < 1024) return `${kb.toFixed(2)} KB`;
  return `${(kb / 1024).toFixed(2)} MB`;
};

const getReportComponent = (report) => {
  const components = {
    'ImageReport': ImageReportCard,
    'SampleReport': SampleReportCard,
    'EnvReport': EnvReportCard,
    'RepairReport': RepairReportCard,
    'TopoReport': TopoReportCard,
    'InstallReport': InstallReportCard
  };

  if (!report) return 'div';

  // First try: instances constructed on the client (have constructor names)
  const typeName = report.constructor && report.constructor.name;
  if (typeName && components[typeName]) return components[typeName];

  // Fallback: API returns plain objects; use numeric `taskType` field
  const tt = report.taskType;
  switch (tt) {
    case 0:
      return ImageReportCard;
    case 1:
      return SampleReportCard;
    case 2:
      return EnvReportCard;
    case 3:
      return RepairReportCard;
    case 4:
      return TopoReportCard;
    case 5:
      return InstallReportCard;
    default:
      return 'div';
  }
};
</script>

<style scoped>
.mission-detail {
  animation: fadeIn 0.2s ease-in;
}

.detail-header {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
  margin-bottom: 24px;
}

.mission-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.mission-title h2 {
  color: var(--text-primary);
  font-size: 22px;
  font-weight: 600;
  margin: 0;
}

.state-badge {
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
}

.state-badge.pending,
.state-badge.Pending {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.state-badge.moving-to,
.state-badge.moving\ to,
.state-badge.Moving\ to {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.state-badge.in\ progress,
.state-badge.In\ Progress,
.state-badge.inprogress,
.state-badge.InProgress,
.state-badge.in-progress {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-primary);
}

.state-badge.completed,
.state-badge.Completed {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}

.mission-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-label {
  color: var(--text-secondary);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.meta-value {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

.meta-value.priority {
  display: inline-block;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  width: fit-content;
}

.meta-value.priority.p-1 {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}

.meta-value.priority.p-2 {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.meta-value.priority.p-3 {
  background: rgba(59, 130, 246, 0.15);
  color: var(--accent-primary);
}

/* Reports Section */
.reports-section {
  margin-top: 24px;
}

.reports-section h3 {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

/* Assembled Image Section */
.assembled-image-section {
  margin-top: 24px;
  margin-bottom: 24px;
}

.assembled-image-section h3 {
  color: var(--accent-success);
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.assembled-image-container {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  text-align: center;
}

.assembled-image-container img {
  max-width: 100%;
  max-height: 600px;
  border-radius: var(--radius-sm);
}

.image-info {
  margin-top: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.debug-info {
  margin-top: 16px;
  padding: 12px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: var(--radius-sm);
  color: var(--accent-warning);
  text-align: center;
  font-size: 13px;
}

.no-reports {
  text-align: center;
  padding: 32px 16px;
  color: var(--text-secondary);
  background: var(--bg-primary);
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-md);
}

.reports-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@media (max-width: 768px) {
  .mission-meta {
    grid-template-columns: 1fr;
  }

  .reports-list {
    grid-template-columns: 1fr;
  }
}
</style>
