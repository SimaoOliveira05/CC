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
  background: var(--bg-primary);
  border-radius: var(--radius-md);
  padding: 14px;
  border-left: 3px solid var(--accent-danger);
  transition: background 0.2s;
}

.report-card.success {
  border-left-color: var(--accent-success);
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
  color: var(--accent-danger);
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.report-card.success .report-header h4 {
  color: var(--accent-success);
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

.status-display {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 600;
}

.status-icon.success {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
  border: 2px solid var(--accent-success);
}

.status-icon.failed {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
  border: 2px solid var(--accent-danger);
}

.status-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  color: var(--text-secondary);
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.status {
  font-weight: 500;
  font-size: 13px;
}

.status.success {
  color: var(--accent-success);
}

.status.failed {
  color: var(--accent-danger);
}
</style>
