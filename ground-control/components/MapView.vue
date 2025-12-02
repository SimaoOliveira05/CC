<template>
  <div class="map-view">
    <h2>Mapa de Rovers e Missões</h2>
    <div class="map-info">
      <span>Rovers: {{ rovers.length }}</span>
      <span>Missões: {{ missions.length }}</span>
    </div>
    <div class="map-container">
      <svg class="map-svg" viewBox="-1.2 -1.2 2.4 2.4" preserveAspectRatio="xMidYMid meet">
          <!-- Euclidean square boundary -->
          <rect x="-1" y="-1" width="2" height="2"
            fill="rgba(0, 0, 0, 0.3)"
            stroke="#0af"
            stroke-width="0.02" />

          <!-- Grid circles (concentric) -->
          <g class="grid">
            <!-- Square grid lines -->
            <line v-for="i in 4" :key="'h' + i"
              x1="-1" :y1="-1 + i * 0.5" x2="1" :y2="-1 + i * 0.5"
              stroke="rgba(0, 212, 255, 0.1)" stroke-width="0.005"/>
            <line v-for="i in 4" :key="'v' + i"
              :x1="-1 + i * 0.5" y1="-1" :x2="-1 + i * 0.5" y2="1"
              stroke="rgba(0, 212, 255, 0.1)" stroke-width="0.005"/>

            <!-- Axes -->
            <line x1="-1" y1="0" x2="1" y2="0" stroke="rgba(0, 212, 255, 0.3)" stroke-width="0.01"/>
            <line x1="0" y1="-1" x2="0" y2="1" stroke="rgba(0, 212, 255, 0.3)" stroke-width="0.01"/>
          </g>

          <!-- Plot labels first, markers last to avoid overlap -->
          <g class="labels">
            <g v-for="mission in validMissions" :key="'ml' + mission.id">
              <text 
                :x="toX(mission.coordinate)" 
                :y="toY(mission.coordinate) - 0.05"
                text-anchor="middle"
                fill="#0af"
                font-size="0.06"
                font-weight="600"
                class="mission-label"
              >
                M{{ mission.id }}
              </text>
            </g>
            <g v-for="rover in validRovers" :key="'rl' + rover.id">
              <text 
                :x="toX(rover.position)" 
                :y="toY(rover.position) - 0.05"
                text-anchor="middle"
                fill="#fff"
                font-size="0.05"
                font-weight="700"
                class="rover-label"
              >
                R{{ rover.id }}
              </text>
            </g>
          </g>

          <!-- Plot markers last for visual priority -->
          <g class="missions">
            <g v-for="mission in validMissions" :key="'m' + mission.id">
              <rect 
                :x="toX(mission.coordinate) - 0.02" 
                :y="toY(mission.coordinate) - 0.02"
                width="0.04" height="0.04"
                :fill="getMissionColor(mission.state)"
                :stroke="getMissionStroke(mission.state)"
                stroke-width="0.012"
                class="mission-marker"
              />
            </g>
          </g>

          <g class="rovers">
            <g v-for="rover in validRovers" :key="'r' + rover.id">
              <circle 
                :cx="toX(rover.position)" 
                :cy="toY(rover.position)"
                :r="0.035"
                fill="#00c853"
                stroke="#1de9b6"
                stroke-width="0.02"
                class="rover-marker"
              />
            </g>
          </g>
      </svg>
    </div>

    <!-- Legend -->
    <div class="legend">
      <div class="legend-item">
        <div class="legend-marker rover-marker-legend"></div>
        <span>Rovers</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker mission-pending"></div>
        <span>Missão Pendente</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker mission-moving"></div>
        <span>Missão Moving to</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker mission-progress"></div>
        <span>Missão In Progress</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker mission-completed"></div>
        <span>Missão Completa</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  rovers: {
    type: Array,
    default: () => []
  },
  missions: {
    type: Array,
    default: () => []
  }
});

// Filtrar rovers e missões com coordenadas válidas
const validRovers = computed(() => {
  return props.rovers.filter(r => 
    r.position && 
    r.position.latitude !== undefined && 
    r.position.longitude !== undefined &&
    r.position.latitude !== 0 &&
    r.position.longitude !== 0
  );
});

const validMissions = computed(() => {
  return props.missions.filter(m => 
    m.coordinate && 
    m.coordinate.latitude !== undefined && 
    m.coordinate.longitude !== undefined
  );
});

// Simplesmente usar as coordenadas lat/lon diretamente no círculo
// O backend usa Haversine para distâncias reais, mas visualmente mostramos num plano
const toX = (coord) => {
  if (!coord) return 0;
  return coord.longitude; // longitude é X
};

const toY = (coord) => {
  if (!coord) return 0;
  // Invert Y to match screen coordinates (SVG Y grows down)
  return -coord.latitude; 
};

const getMissionColor = (state) => {
  if (!state) return '#ffaa00';
  const s = state.toLowerCase();
  if (s === 'completed') return '#00ff88';
  if (s === 'in progress' || s === 'in-progress') return '#ff4444';
  if (s === 'moving to' || s === 'moving-to') return '#ff9500';
  return '#ffaa00'; // pending
};

const getMissionStroke = (state) => {
  if (!state) return '#ffcc00';
  const s = state.toLowerCase();
  if (s === 'completed') return '#00ffaa';
  if (s === 'in progress' || s === 'in-progress') return '#ff6666';
  if (s === 'moving to' || s === 'moving-to') return '#ffaa00';
  return '#ffcc00'; // pending
};
</script>

<style scoped>
.map-view {
  animation: fadeIn 0.3s ease-in;
}

.map-view h2 {
  color: #00d4ff;
  margin-bottom: 20px;
  font-size: 24px;
  text-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.map-info {
  display: flex;
  gap: 20px;
  margin-bottom: 15px;
  color: #a8b5c8;
  font-size: 14px;
}

.map-info span {
  padding: 5px 10px;
  background: rgba(0, 212, 255, 0.1);
  border-radius: 4px;
  border: 1px solid rgba(0, 212, 255, 0.3);
}

.map-container {
  background: linear-gradient(135deg, #0a1929 0%, #132d48 100%);
  border: 2px solid #00d4ff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
  margin-bottom: 20px;
}

.map-svg {
  width: 100%;
  height: 600px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 4px;
}

.rover-marker {
  cursor: pointer;
  transition: all 0.3s;
  filter: drop-shadow(0 0 5px rgba(0, 255, 136, 0.6));
}

.rover-marker:hover {
  filter: drop-shadow(0 0 10px rgba(0, 255, 136, 1));
  transform: scale(1.2);
}

.mission-marker {
  cursor: pointer;
  transition: all 0.3s;
  opacity: 0.8;
}

.mission-marker:hover {
  opacity: 1;
  transform: scale(1.2);
}

.rover-label,
.mission-label {
  pointer-events: none;
  user-select: none;
}

.legend {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  justify-content: center;
  padding: 15px;
  background: rgba(0, 212, 255, 0.05);
  border-radius: 8px;
  border: 1px solid rgba(0, 212, 255, 0.2);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e8eef7;
  font-size: 14px;
}

.legend-marker {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid;
}

.rover-marker-legend {
  background: #00ff88;
  border-color: #00ffaa;
}

.mission-pending {
  background: rgba(255, 170, 0, 0.6);
  border-color: #ffcc00;
}

.mission-moving {
  background: rgba(255, 149, 0, 0.6);
  border-color: #ffaa00;
}

.mission-progress {
  background: rgba(255, 68, 68, 0.6);
  border-color: #ff6666;
}

.mission-completed {
  background: rgba(0, 255, 136, 0.6);
  border-color: #00ffaa;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
