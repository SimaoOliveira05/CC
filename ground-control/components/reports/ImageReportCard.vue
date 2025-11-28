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
  background: linear-gradient(135deg, #1a3a52 0%, #132d48 100%);
  border-radius: 8px;
  padding: 15px;
  border-left: 4px solid #00d4ff;
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.1);
  transition: all 0.3s;
}

.report-card:hover {
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3);
  transform: translateY(-2px);
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(0, 212, 255, 0.2);
}

.report-header h4 {
  color: #00d4ff;
  margin: 0;
  font-size: 14px;
  text-shadow: 0 0 5px rgba(0, 212, 255, 0.3);
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
  gap: 8px;
}

.info-item {
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
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
}

.status.partial {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
}

.status.complete {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
}
</style>
