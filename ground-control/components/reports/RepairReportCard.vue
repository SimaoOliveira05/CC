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
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #ff6b1f;
  box-shadow: 0 0 10px rgba(255, 107, 31, 0.1);
  transition: all 0.3s;
}

.report-card.success {
  border-left-color: #00ff88;
  box-shadow: 0 0 10px rgba(0, 255, 136, 0.1);
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(255, 107, 31, 0.3);
  transform: translateY(-2px);
}

.report-card.success:hover {
  box-shadow: 0 0 15px rgba(0, 255, 136, 0.3);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(255, 107, 31, 0.2);
}

.report-card.success .report-header {
  border-bottom-color: rgba(0, 255, 136, 0.2);
}

.report-header h4 {
  color: #ff6b1f;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(255, 107, 31, 0.3);
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
  gap: 10px;
}

.info-item,
.status-item {
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

.status {
  padding: 4px 8px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
}

.status.repairable {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
}

.status.not-repairable {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
}
</style>
