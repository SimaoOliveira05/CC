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
  animation: slideIn 0.3s ease-out;
}

.detail-header {
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border: 2px solid #00d4ff;
  border-radius: 8px;
  padding: 25px;
  margin-bottom: 30px;
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
}

.mission-title {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.mission-title h2 {
  color: #00d4ff;
  font-size: 28px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
  margin: 0;
}

.state-badge {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
}

/* Pending states */
.state-badge.pending,
.state-badge.Pending {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
  border: 1px solid #ffaa00;
}

/* Moving to states */
.state-badge.moving-to,
.state-badge.moving\ to,
.state-badge.Moving\ to {
  background: rgba(255, 170, 0, 0.2);
  color: #ff9500;
  border: 1px solid #ff9500;
}

/* In Progress states */
.state-badge.in\ progress,
.state-badge.In\ Progress,
.state-badge.inprogress,
.state-badge.InProgress,
.state-badge.in-progress {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
  border: 1px solid #ff4444;
}

/* Completed states */
.state-badge.completed,
.state-badge.Completed {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  border: 1px solid #00ff88;
}

.mission-meta {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.meta-label {
  color: #a8b5c8;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.meta-value {
  color: #e8eef7;
  font-size: 14px;
  font-weight: bold;
}

.meta-value.priority {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 13px;
}

.meta-value.priority.p-1 {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
}

.meta-value.priority.p-2 {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
}

.meta-value.priority.p-3 {
  background: rgba(0, 212, 255, 0.2);
  color: #00d4ff;
}

/* Reports Section */
.reports-section {
  margin-top: 30px;
}

.reports-section h3 {
  color: #00d4ff;
  font-size: 20px;
  margin-bottom: 20px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
  text-transform: uppercase;
  letter-spacing: 1px;
}

/* Assembled Image Section */
.assembled-image-section {
  margin-top: 30px;
  margin-bottom: 30px;
}

.assembled-image-section h3 {
  color: #00ff88;
  font-size: 20px;
  margin-bottom: 20px;
  text-shadow: 0 0 10px rgba(0, 255, 136, 0.3);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.assembled-image-container {
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border: 2px solid #00ff88;
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  box-shadow: 0 0 20px rgba(0, 255, 136, 0.2);
}

.assembled-image-container img {
  max-width: 100%;
  max-height: 600px;
  border-radius: 4px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.5);
  image-rendering: auto;
}

.image-info {
  margin-top: 15px;
  color: #a8b5c8;
  font-size: 14px;
}

.debug-info {
  margin-top: 20px;
  padding: 15px;
  background: rgba(255, 170, 0, 0.1);
  border: 1px solid rgba(255, 170, 0, 0.3);
  border-radius: 6px;
  color: #ffaa00;
  text-align: center;
  font-size: 13px;
}

.no-reports {
  text-align: center;
  padding: 40px 20px;
  color: #a8b5c8;
  background: rgba(10, 30, 61, 0.5);
  border: 2px dashed #1a3a52;
  border-radius: 8px;
}

.reports-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
