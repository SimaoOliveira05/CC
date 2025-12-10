<template>
  <div class="app">
    <!-- Header NASA -->
    <header class="header">
      <div class="header-content">
        <div class="logo">
          <span class="nasa-badge">🚀</span>
          <h1>NasUM Ground Control</h1>
        </div>
        <div class="status-indicator">
          <span class="pulse" :class="apiConnected ? 'connected' : 'disconnected'"></span>
          {{ apiConnected ? 'Conectado' : 'Desconectado' }}
        </div>
        <div class="last-update" v-if="lastUpdate">
          Último update: {{ lastUpdate }}
        </div>
      </div>
    </header>

    <!-- Main Container -->
    <div class="container">
      <!-- Sidebar - Rovers -->
      <aside class="sidebar">
        <h2>Rovers Ativos</h2>
        <div class="rovers-grid">
          <RoverCard 
            v-for="rover in sortedRovers" 
            :key="rover.id" 
            :rover="rover"
          />
        </div>
      </aside>

      <!-- Main Content -->
      <main class="main-content">
        <!-- Tab Navigation -->
        <div class="tabs">
          <button 
            class="tab" 
            :class="{ active: activeTab === 'missions' }"
            @click="activeTab = 'missions'"
          >
            📋 Missões
          </button>
          <button 
            class="tab" 
            :class="{ active: activeTab === 'map' }"
            @click="activeTab = 'map'"
          >
            🗺️ Mapa
          </button>
          <button 
            class="tab" 
            :class="{ active: activeTab === 'logs' }"
            @click="activeTab = 'logs'"
          >
            📝 Logs
            <span v-if="unreadLogs > 0" class="log-badge">{{ unreadLogs }}</span>
          </button>
        </div>

        <!-- Missions Tab -->
        <section v-if="activeTab === 'missions' && !selectedMission" class="missions-section">
          <h2>Missões Ativas</h2>
          <div class="missions-grid">
            <MissionCard 
              v-for="mission in activeMissions" 
              :key="mission.id" 
              :mission="mission"
              @select="selectMission"
            />
          </div>
          <div v-if="activeMissions && activeMissions.length === 0" class="empty-state">
            <p>Nenhuma missão ativa no momento</p>
          </div>

          <h2 style="margin-top: 40px;">Missões Completas</h2>
          <div class="missions-grid">
            <MissionCard 
              v-for="mission in completedMissions" 
              :key="mission.id" 
              :mission="mission"
              @select="selectMission"
            />
          </div>
          <div v-if="completedMissions && completedMissions.length === 0" class="empty-state">
            <p>Nenhuma missão completa</p>
          </div>
        </section>

        <!-- Map Tab -->
        <section v-if="activeTab === 'map'" class="map-section">
          <MapView :rovers="sortedRovers" :missions="missions" />
        </section>

        <!-- Logs Tab -->
        <section v-if="activeTab === 'logs'" class="logs-section">
          <LogTab :logs="logs" @clear="clearLogs" />
        </section>

        <!-- Mission Detail -->
        <section v-if="selectedMission" class="mission-detail-section">
          <button class="btn-back" @click="selectedMission = null">← Voltar</button>
          <MissionDetail :mission="selectedMission" />
        </section>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { Rover, Mission } from './models.js';
import RoverCard from './components/RoverCard.vue';
import MissionCard from './components/MissionCard.vue';
import MissionDetail from './components/MissionDetail.vue';
import MapView from './components/MapView.vue';
import LogTab from './components/LogTab.vue';

// Onde defines o URL da API
// Antes tinhas algo como: const API_URL = "http://localhost:8080";

// Agora usa isto:
const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

console.log("Conectando à Nave Mãe em:", API_URL);

// Usa API_URL nas tuas chamadas fetch ou WebSocket
// Exemplo: new WebSocket(API_URL.replace("http", "ws") + "/ws");

const API_BASE = `${API_URL}/api`;
const WS_BASE = `${API_URL.replace("http", "ws")}/ws`; // (Confirma se a rota do WS é /ws ou /ws/telemetry)
const rovers = ref([]);
const missions = ref([]);
const selectedMission = ref(null);
const apiConnected = ref(false);
const lastUpdate = ref(null);
const updateInterval = ref(null);
const ws = ref(null);
const activeTab = ref('missions');
const logs = ref([]);
const unreadLogs = ref(0);

// Limitar número máximo de logs
const MAX_LOGS = 500;

// Computed: Rovers ordenados por ID
const sortedRovers = computed(() => {
  return [...rovers.value].sort((a, b) => a.id - b.id);
});

// Computed: Missões ativas (não completas) ordenadas por ID
const activeMissions = computed(() => {
  return missions.value
    .filter(m => m.state !== 'Completed')
    .sort((a, b) => a.id - b.id); // Ordenadas por ID
});

// Computed: Missões completas ordenadas por ID
const completedMissions = computed(() => {
  return missions.value
    .filter(m => m.state === 'Completed')
    .sort((a, b) => a.id - b.id); // Ordenadas por ID
});

const selectMission = (mission) => {
  selectedMission.value = mission;
};

// Limpar logs
const clearLogs = () => {
  logs.value = [];
  unreadLogs.value = 0;
};

// Adicionar log
const addLog = (logEvent) => {
  logs.value.push(logEvent);
  // Manter apenas os últimos MAX_LOGS
  if (logs.value.length > MAX_LOGS) {
    logs.value = logs.value.slice(-MAX_LOGS);
  }
  // Incrementar contador se não estiver na aba de logs
  if (activeTab.value !== 'logs') {
    unreadLogs.value++;
  }
};

// Reset unread quando muda para aba de logs
watch(activeTab, (newTab) => {
  if (newTab === 'logs') {
    unreadLogs.value = 0;
  }
});

// Carregar dados da API
const loadData = async () => {
  try {
    // Carregar rovers
    const r = await fetch(`${API_BASE}/rovers`);
    if (r.ok) {
      const roverArr = await r.json();
      rovers.value = (roverArr || []).map(obj => new Rover(obj));
      apiConnected.value = true;
    }
  } catch (e) {
    console.error('Erro ao carregar rovers:', e);
    apiConnected.value = false;
  }

  try {
    // Carregar missões
    const m = await fetch(`${API_BASE}/missions`);
    if (m.ok) {
      const missionArr = await m.json();
      missions.value = (missionArr || []).map(obj => new Mission(obj));
      apiConnected.value = true;
    }
  } catch (e) {
    console.error('Erro ao carregar missões:', e);
    apiConnected.value = false;
  }

  // Atualizar timestamp
  lastUpdate.value = new Date().toLocaleTimeString('pt-PT');

  // Se há uma missão selecionada, atualizar também
  if (selectedMission.value) {
    const updated = missions.value.find(m => m.id === selectedMission.value.id);
    if (updated) {
      selectedMission.value = updated;
    }
  }
};

// Conectar WebSocket (se disponível)
const connectWebSocket = () => {
  try {
    ws.value = new WebSocket(`${WS_BASE}/telemetry`);

    ws.value.onopen = () => {
      console.log('WebSocket conectado');
      apiConnected.value = true;
    };

    ws.value.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      console.log('Mensagem WebSocket:', msg);
      // Recarregar dados quando receber update
      if (msg.event === 'snapshot' || msg.event === 'update') {
        loadData();
      }
      // Processar eventos de log
      if (msg.event === 'log' && msg.data) {
        addLog(msg.data);
      }
    };

    ws.value.onerror = (error) => {
      console.error('Erro WebSocket:', error);
    };

    ws.value.onclose = () => {
      console.log('WebSocket desconectado, tentando reconectar...');
      apiConnected.value = false;
      setTimeout(connectWebSocket, 3000); // Tentar reconectar em 3s
    };
  } catch (e) {
    console.error('Erro ao conectar WebSocket:', e);
  }
};

onMounted(() => {
  // Carregar dados inicialmente
  loadData();

  // Tentar WebSocket
  connectWebSocket();

  // Polling a cada 2 segundos como fallback
  updateInterval.value = setInterval(loadData, 2000);
});

onUnmounted(() => {
  // Limpar intervalo
  if (updateInterval.value) {
    clearInterval(updateInterval.value);
  }
  // Fechar WebSocket
  if (ws.value) {
    ws.value.close();
  }
});
</script>

<style scoped>
/* ===== HEADER ===== */
.header {
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  padding: 16px 0;
  position: sticky;
  top: 0;
  z-index: 1000;
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nasa-badge {
  font-size: 28px;
}

.logo h1 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
}

.last-update {
  color: var(--text-secondary);
  font-size: 12px;
}

.pulse {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.pulse.connected {
  background: var(--accent-success);
}

.pulse.disconnected {
  background: var(--accent-danger);
}

/* ===== LAYOUT ===== */
.container {
  display: flex;
  max-width: 1400px;
  margin: 20px auto;
  gap: 20px;
  padding: 0 20px;
  min-height: calc(100vh - 100px);
}

.sidebar {
  width: 280px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 20px;
}

.sidebar h2 {
  color: var(--text-primary);
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
}

.rovers-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.main-content {
  flex: 1;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 24px;
}

.main-content h2 {
  color: var(--text-primary);
  margin-bottom: 20px;
  font-size: 18px;
  font-weight: 600;
}

/* ===== TABS ===== */
.tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0;
}

.tab {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 10px 16px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  position: relative;
}

.tab:hover {
  color: var(--text-primary);
}

.tab.active {
  color: var(--accent-primary);
  border-bottom-color: var(--accent-primary);
}

.log-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  background: var(--accent-danger);
  color: white;
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 16px;
  text-align: center;
}

/* ===== LOGS SECTION ===== */
.logs-section {
  animation: fadeIn 0.2s ease-in;
  height: calc(100vh - 250px);
  min-height: 400px;
}

/* ===== MISSIONS SECTION ===== */
.missions-section {
  animation: fadeIn 0.2s ease-in;
}

.missions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.empty-state {
  text-align: center;
  padding: 48px 20px;
  color: var(--text-secondary);
}

.empty-state p {
  font-size: 15px;
}

/* ===== MISSION DETAIL SECTION ===== */
.mission-detail-section {
  animation: fadeIn 0.2s ease-in;
}

.btn-back {
  background: var(--bg-hover);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  margin-bottom: 16px;
  transition: all 0.2s;
  font-size: 13px;
  font-weight: 500;
}

.btn-back:hover {
  background: var(--border-color);
}

/* ===== ANIMATIONS ===== */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* ===== RESPONSIVE ===== */
@media (max-width: 1024px) {
  .container {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
  }

  .rovers-grid {
    flex-direction: row;
    overflow-x: auto;
    gap: 12px;
  }

  .missions-grid {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }
}
</style>
