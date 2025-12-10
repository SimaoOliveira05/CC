<template>
  <div class="report-card image-report">
    <div class="report-header">
      <h4>📸 Captura de Imagem</h4>
      <span v-if="report.isLastReport" class="last-badge">Último</span>
    </div>
    <div class="report-body">
      <div class="info-item">
        <span class="label">Chunk ID:</span>
        <span class="value">#{{ report.chunkId }}</span>
      </div>
      <div class="info-item">
        <span class="label">Tamanho:</span>
        <span class="value">{{ formatBytes(report.data?.length || 0) }}</span>
      </div>
      <div class="info-item">
        <span class="label">Status:</span>
        <span class="status" :class="report.isLastReport ? 'complete' : 'partial'">
          {{ report.isLastReport ? 'Completo' : 'Parcial' }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue';

const props = defineProps({
  report: Object,
  required: true
});

const imageUrl = ref(null);

const updateImageUrl = (data) => {
  // Revoke previous URL if any
  if (imageUrl.value && imageUrl.value.startsWith && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value);
  }
  if (!data) {
    imageUrl.value = null;
    return;
  }
  // If assembled image is present on the report, prefer it
  const assembled = props.report && (props.report.assembledImage || props.report.assembledImageBase64);
  if (assembled) {
    imageUrl.value = `data:image/jpeg;base64,${assembled}`;
    return;
  }
  if (typeof data === 'string') {
    // already base64
    imageUrl.value = `data:image/jpeg;base64,${data}`;
    return;
  }
  try {
    // data is expected to be Uint8Array or ArrayBuffer-like
    const blob = new Blob([data], { type: 'image/jpeg' });
    imageUrl.value = URL.createObjectURL(blob);
  } catch (e) {
    imageUrl.value = null;
  }
};

watch(() => props.report && props.report.data, (newData) => {
  updateImageUrl(newData);
}, { immediate: true });

onUnmounted(() => {
  if (imageUrl.value && imageUrl.value.startsWith && imageUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(imageUrl.value);
  }
});

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};
</script>

<style scoped>
.report-card {
  background: var(--bg-primary);
  border-radius: var(--radius-md);
  padding: 14px;
  border-left: 3px solid var(--accent-primary);
  transition: background 0.2s;
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
  color: var(--accent-primary);
  margin: 0;
  font-size: 13px;
  font-weight: 600;
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
  gap: 6px;
}

.info-item {
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
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 500;
  text-transform: uppercase;
}

.status.partial {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.status.complete {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}
</style>
