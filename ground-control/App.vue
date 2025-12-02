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
/* ===== VARIABLES ===== */
:root {
  --primary-dark: #0a1e3d;
  --primary-blue: #1e5a96;
  --accent-cyan: #00d4ff;
  --accent-green: #00ff88;
  --accent-orange: #ff6b1f;
  --text-primary: #e8eef7;
  --text-secondary: #a8b5c8;
  --border-color: #1a3a52;
  --card-bg: #0f2440;
  --input-bg: #132d48;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body, html {
  background: linear-gradient(135deg, #0a1e3d 0%, #132d48 100%);
  color: var(--text-primary);
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

/* ===== HEADER ===== */
.header {
  background: linear-gradient(90deg, #0a1e3d 0%, #1e5a96 100%);
  border-bottom: 3px solid var(--accent-cyan);
  padding: 20px 0;
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.3);
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

.nasa-badge {
  font-size: 32px;
}

.logo h1 {
  font-size: 28px;
  color: var(--accent-cyan);
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid var(--accent-cyan);
  border-radius: 4px;
  color: var(--accent-green);
}

.last-update {
  color: var(--accent-cyan);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
  animation: pulse-text 2s infinite;
}

@keyframes pulse-text {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

.pulse {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.pulse.connected {
  background: var(--accent-green);
  box-shadow: 0 0 10px var(--accent-green);
}

.pulse.disconnected {
  background: #ff4444;
  box-shadow: 0 0 10px #ff4444;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
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
  width: 250px;
  background: rgba(15, 36, 64, 0.8);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 20px;
  backdrop-filter: blur(10px);
}

.sidebar h2 {
  color: var(--accent-cyan);
  margin-bottom: 20px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
  font-size: 18px;
}

.rovers-grid {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.main-content {
  flex: 1;
  background: rgba(15, 36, 64, 0.8);
  border: 2px solid var(--border-color);
  border-radius: 8px;
  padding: 30px;
  backdrop-filter: blur(10px);
}

.main-content h2 {
  color: var(--accent-cyan);
  margin-bottom: 25px;
  font-size: 24px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

/* ===== TABS ===== */
.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 30px;
  border-bottom: 2px solid rgba(0, 212, 255, 0.2);
  padding-bottom: 10px;
}

.tab {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 10px 20px;
  border-radius: 4px 4px 0 0;
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
  transition: all 0.3s;
  border-bottom: 3px solid transparent;
}

.tab:hover {
  color: var(--accent-cyan);
  background: rgba(0, 212, 255, 0.1);
}

.tab.active {
  color: var(--accent-cyan);
  background: rgba(0, 212, 255, 0.15);
  border-bottom-color: var(--accent-cyan);
}

.tab {
  position: relative;
}

.log-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  background: #ff4444;
  color: white;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

/* ===== LOGS SECTION ===== */
.logs-section {
  animation: fadeIn 0.3s ease-in;
  height: calc(100vh - 250px);
  min-height: 400px;
}

/* ===== MISSIONS SECTION ===== */
.missions-section {
  animation: fadeIn 0.3s ease-in;
}

.missions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.empty-state p {
  font-size: 18px;
}

/* ===== MISSION DETAIL SECTION ===== */
.mission-detail-section {
  animation: slideIn 0.3s ease-out;
}

.btn-back {
  background: var(--accent-orange);
  border: none;
  color: white;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  margin-bottom: 20px;
  transition: all 0.3s;
  font-weight: bold;
}

.btn-back:hover {
  background: #ff8844;
  box-shadow: 0 0 15px rgba(255, 107, 31, 0.5);
}

/* ===== ANIMATIONS ===== */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
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
  }

  .missions-grid {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }
}
</style>
