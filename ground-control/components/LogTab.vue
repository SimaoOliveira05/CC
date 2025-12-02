<template>
  <div class="logs-container">
    <div class="logs-header">
      <h2>📝 Logs do Sistema</h2>
      <div class="logs-controls">
        <button @click="clearLogs" class="btn-clear">🗑️ Limpar</button>
      </div>
    </div>
    
    <div class="logs-list" ref="logsContainer">
      <div 
        v-for="(log, index) in props.logs" 
        :key="index" 
        class="log-entry"
        :class="log.level.toLowerCase()"
      >
        <span class="log-time">{{ formatTime(log.timestamp) }}</span>
        <span class="log-level" :class="log.level.toLowerCase()">{{ log.level }}</span>
        <span class="log-source">{{ log.source }}</span>
        <span class="log-message">{{ log.message }}</span>
        <span v-if="log.meta" class="log-meta">{{ formatMeta(log.meta) }}</span>
      </div>
      
      <div v-if="props.logs.length === 0" class="empty-logs">
        <p>Nenhum log disponível</p>
      </div>
    </div>
    
    <div class="logs-footer">
      <span class="log-count">{{ props.logs.length }} eventos</span>
      <label class="auto-scroll-toggle">
        <input type="checkbox" v-model="autoScroll" />
        Auto-scroll
      </label>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue';

const props = defineProps({
  logs: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['clear']);

const autoScroll = ref(true);
const logsContainer = ref(null);

const formatTime = (timestamp) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('pt-PT', { 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit',
    fractionalSecondDigits: 3
  });
};

const formatMeta = (meta) => {
  if (typeof meta === 'object') {
    return JSON.stringify(meta);
  }
  return String(meta);
};

const clearLogs = () => {
  emit('clear');
};

// Auto-scroll quando novos logs chegam
watch(() => props.logs.length, async () => {
  if (autoScroll.value && logsContainer.value) {
    await nextTick();
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight;
  }
});
</script>

<style scoped>
.logs-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 500px;
}

.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 15px;
}

.logs-header h2 {
  color: var(--accent-cyan, #00d4ff);
  margin: 0;
  font-size: 24px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.logs-controls {
  display: flex;
  gap: 10px;
  align-items: center;
}

.btn-clear {
  background: rgba(255, 68, 68, 0.2);
  border: 1px solid #ff4444;
  color: #ff4444;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-clear:hover {
  background: rgba(255, 68, 68, 0.4);
}

.logs-list {
  flex: 1;
  background: rgba(10, 30, 61, 0.6);
  border: 1px solid rgba(26, 58, 82, 0.8);
  border-radius: 8px;
  padding: 15px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px 10px;
  border-radius: 4px;
  margin-bottom: 4px;
  transition: background 0.2s;
}

.log-entry:hover {
  background: rgba(0, 212, 255, 0.05);
}

.log-entry.info {
  border-left: 3px solid #00ff88;
}

.log-entry.warn {
  border-left: 3px solid #ffaa00;
}

.log-entry.error {
  border-left: 3px solid #ff4444;
}

.log-time {
  color: #a8b5c8;
  font-size: 12px;
  min-width: 100px;
}

.log-level {
  font-weight: bold;
  min-width: 50px;
  text-align: center;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
}

.log-level.info {
  background: rgba(0, 255, 136, 0.2);
  color: #00ff88;
}

.log-level.warn {
  background: rgba(255, 170, 0, 0.2);
  color: #ffaa00;
}

.log-level.error {
  background: rgba(255, 68, 68, 0.2);
  color: #ff4444;
}

.log-source {
  color: #00d4ff;
  min-width: 120px;
  font-size: 12px;
}

.log-message {
  color: #e8eef7;
  flex: 1;
}

.log-meta {
  color: #ff6b1f;
  font-size: 11px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-logs {
  text-align: center;
  color: #a8b5c8;
  padding: 60px 20px;
}

.logs-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid rgba(26, 58, 82, 0.5);
}

.log-count {
  color: #a8b5c8;
  font-size: 13px;
}

.auto-scroll-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #a8b5c8;
  font-size: 13px;
  cursor: pointer;
}

.auto-scroll-toggle input {
  accent-color: #00d4ff;
}
</style>
