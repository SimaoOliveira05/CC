<template>
  <div class="report-card install-report" :class="report.success ? 'success' : 'failed'">
    <div class="report-header">
      <h4>⚙️ Instalação</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="status-display">
        <div class="status-icon" :class="report.success ? 'success' : 'failed'">
          {{ report.success ? '✓' : '✗' }}
        </div>
        <div class="status-text">
          <span class="label">Status da Instalação</span>
          <span class="status" :class="report.success ? 'success' : 'failed'">
            {{ report.success ? 'Concluída com Sucesso' : 'Falhou' }}
          </span>
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
  border-left: 4px solid #ff4444;
  box-shadow: 0 0 10px rgba(255, 68, 68, 0.1);
  transition: all 0.3s;
}

.report-card.success {
  border-left-color: #00ff88;
  box-shadow: 0 0 10px rgba(0, 255, 136, 0.1);
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(255, 68, 68, 0.3);
  transform: translateY(-2px);
}

.report-card.success:hover {
  box-shadow: 0 0 15px rgba(0, 255, 136, 0.3);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 68, 68, 0.2);
}

.report-card.success .report-header {
  border-bottom-color: rgba(0, 255, 136, 0.2);
}

.report-header h4 {
  color: #ff4444;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(255, 68, 68, 0.3);
}

.report-card.success .report-header h4 {
  color: #00ff88;
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

.status-display {
  display: flex;
  align-items: center;
  gap: 15px;
}

.status-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: bold;
}

.status-icon.success {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
  border: 2px solid #00ff88;
  box-shadow: 0 0 10px rgba(0, 255, 136, 0.3);
}

.status-icon.failed {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
  border: 2px solid #ff4444;
  box-shadow: 0 0 10px rgba(255, 68, 68, 0.3);
}

.status-text {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.label {
  color: #a8b5c8;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status {
  font-weight: bold;
  font-size: 14px;
}

.status.success {
  color: #00ff88;
  text-shadow: 0 0 5px rgba(0, 255, 136, 0.3);
}

.status.failed {
  color: #ff4444;
  text-shadow: 0 0 5px rgba(255, 68, 68, 0.3);
}
</style>
