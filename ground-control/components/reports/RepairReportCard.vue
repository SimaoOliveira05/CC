<template>
  <div class="report-card repair-report" :class="report.repairable ? 'success' : 'failed'">
    <div class="report-header">
      <h4>🔧 Reparação/Resgate</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="info-item">
        <span class="label">Problema ID:</span>
        <span class="value">#{{ report.problemId }}</span>
      </div>
      <div class="status-item">
        <span class="label">Status:</span>
        <span class="status" :class="report.repairable ? 'repairable' : 'not-repairable'">
          {{ report.repairable ? '✓ Reparável' : '✗ Não Reparável' }}
        </span>
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
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

.report-header h4 {
  color: var(--accent-warning);
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
  gap: 8px;
}

.info-item,
.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.label {
  color: var(--text-secondary);
  text-transform: uppercase;
  font-size: 10px;
  letter-spacing: 0.3px;
}

.value {
  color: var(--text-primary);
  font-weight: 500;
}

.status {
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 500;
  text-transform: uppercase;
}

.status.repairable {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}

.status.not-repairable {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}
</style>
