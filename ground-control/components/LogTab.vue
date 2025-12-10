<template>
  <div class="logs-container">
    <div class="logs-header">
      <h2>📝 Logs do Sistema</h2>
      <div class="logs-controls">
        <div class="filter-group">
          <span class="filter-label">Filtrar:</span>
          <button 
            @click="toggleFilter('all')" 
            class="filter-btn"
            :class="{ active: activeFilter === 'all' }"
          >
            Todos
          </button>
          <button 
            @click="toggleFilter('info')" 
            class="filter-btn info"
            :class="{ active: activeFilter === 'info' }"
          >
            ℹ️ Info
          </button>
          <button 
            @click="toggleFilter('warn')" 
            class="filter-btn warn"
            :class="{ active: activeFilter === 'warn' }"
          >
            ⚠️ Warn
          </button>
          <button 
            @click="toggleFilter('error')" 
            class="filter-btn error"
            :class="{ active: activeFilter === 'error' }"
          >
            ❌ Error
          </button>
        </div>
        <button @click="clearLogs" class="btn-clear">🗑️ Limpar</button>
      </div>
    </div>
    
    <div class="logs-list" ref="logsContainer">
      <div 
        v-for="(log, index) in filteredLogs" 
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
      
      <div v-if="filteredLogs.length === 0" class="empty-logs">
        <p v-if="props.logs.length === 0">Nenhum log disponível</p>
        <p v-else>Nenhum log encontrado para o filtro "{{ activeFilter }}"</p>
      </div>
    </div>
    
    <div class="logs-footer">
      <span class="log-count">{{ filteredLogs.length }} de {{ props.logs.length }} eventos</span>
      <label class="auto-scroll-toggle">
        <input type="checkbox" v-model="autoScroll" />
        Auto-scroll
      </label>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, computed } from 'vue';

const props = defineProps({
  logs: {
    type: Array,
    default: () => []
  }
});

const emit = defineEmits(['clear']);

const autoScroll = ref(true);
const logsContainer = ref(null);
const activeFilter = ref('all');

const filteredLogs = computed(() => {
  if (activeFilter.value === 'all') {
    return props.logs;
  }
  return props.logs.filter(log => log.level.toLowerCase() === activeFilter.value);
});

const toggleFilter = (filter) => {
  activeFilter.value = filter;
};

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
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
}

.logs-header h2 {
  color: var(--text-primary);
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.logs-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  gap: 6px;
  align-items: center;
}

.filter-label {
  color: var(--text-secondary);
  font-size: 12px;
  margin-right: 4px;
}

.filter-btn {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  padding: 5px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.filter-btn:hover {
  background: var(--bg-hover);
}

.filter-btn.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: var(--accent-primary);
  color: var(--accent-primary);
}

.filter-btn.info.active {
  background: rgba(34, 197, 94, 0.15);
  border-color: var(--accent-success);
  color: var(--accent-success);
}

.filter-btn.warn.active {
  background: rgba(245, 158, 11, 0.15);
  border-color: var(--accent-warning);
  color: var(--accent-warning);
}

.filter-btn.error.active {
  background: rgba(239, 68, 68, 0.15);
  border-color: var(--accent-danger);
  color: var(--accent-danger);
}

.btn-clear {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid var(--accent-danger);
  color: var(--accent-danger);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}

.btn-clear:hover {
  background: rgba(239, 68, 68, 0.2);
}

.logs-list {
  flex: 1;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 12px;
  overflow-y: auto;
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 12px;
}

.log-entry {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 6px 8px;
  border-radius: var(--radius-sm);
  margin-bottom: 2px;
  transition: background 0.15s;
}

.log-entry:hover {
  background: var(--bg-hover);
}

.log-entry.info {
  border-left: 2px solid var(--accent-success);
}

.log-entry.warn {
  border-left: 2px solid var(--accent-warning);
}

.log-entry.error {
  border-left: 2px solid var(--accent-danger);
}

.log-time {
  color: var(--text-muted);
  font-size: 11px;
  min-width: 90px;
}

.log-level {
  font-weight: 500;
  min-width: 44px;
  text-align: center;
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 10px;
}

.log-level.info {
  background: rgba(34, 197, 94, 0.15);
  color: var(--accent-success);
}

.log-level.warn {
  background: rgba(245, 158, 11, 0.15);
  color: var(--accent-warning);
}

.log-level.error {
  background: rgba(239, 68, 68, 0.15);
  color: var(--accent-danger);
}

.log-source {
  color: var(--accent-primary);
  min-width: 100px;
  font-size: 11px;
}

.log-message {
  color: var(--text-primary);
  flex: 1;
}

.log-meta {
  color: var(--accent-warning);
  font-size: 10px;
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-logs {
  text-align: center;
  color: var(--text-secondary);
  padding: 48px 16px;
}

.logs-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.log-count {
  color: var(--text-secondary);
  font-size: 12px;
}

.auto-scroll-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
}

.auto-scroll-toggle input {
  accent-color: var(--accent-primary);
}
</style>
