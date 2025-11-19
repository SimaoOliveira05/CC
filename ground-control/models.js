// Classe Rover
export class Rover {
  constructor({ id, state, battery, speed }) {
    this.id = id;
    this.state = state;
    this.battery = battery;
    this.speed = speed;
  }
}

// Classe Mission
export class Mission {
  constructor({ id, idRover, taskType, duration, updateFrequency, lastUpdate, createdAt, priority, reports, state }) {
    this.id = id;
    this.idRover = idRover;
    this.taskType = taskType;
    this.duration = duration;
    this.updateFrequency = updateFrequency;
    this.lastUpdate = lastUpdate;
    this.createdAt = createdAt;
    this.priority = priority;
    this.reports = (reports || []).map(r => this.instantiateReport(r));
    this.state = state;
  }

  instantiateReport(data) {
    if (!data || !data.taskType) return null;
    
    switch(data.taskType) {
      case 0:
        return new ImageReport(data);
      case 1:
        return new SampleReport(data);
      case 2:
        return new EnvReport(data);
      case 3:
        return new RepairReport(data);
      case 4:
        return new TopoReport(data);
      case 5:
        return new InstallReport(data);
      default:
        return null;
    }
  }
}

// ===== REPORTS =====

// Classe ImageReport
export class ImageReport {
  constructor({ taskType, missionId, chunkId, data, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.chunkId = chunkId;
    this.data = data;
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Imagem';
  }

  getSummary() {
    const dataSize = this.data ? this.data.length : 0;
    return `Chunk #${this.chunkId} (${dataSize} bytes) ${this.isLastReport ? '✓ Último' : ''}`;
  }
}

// Classe SampleReport
export class SampleReport {
  constructor({ taskType, missionId, numSamples, components, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.numSamples = numSamples;
    this.components = components || [];
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Amostra';
  }

  getSummary() {
    const compStr = this.components.map(c => `${c.name}=${c.percentage.toFixed(2)}%`).join(', ');
    return `${this.numSamples} componentes: [${compStr}] ${this.isLastReport ? '✓ Último' : ''}`;
  }
}

// Classe EnvReport
export class EnvReport {
  constructor({ taskType, missionId, temp, oxygen, pressure, humidity, windSpeed, radiation, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.temp = temp;
    this.oxygen = oxygen;
    this.pressure = pressure;
    this.humidity = humidity;
    this.windSpeed = windSpeed;
    this.radiation = radiation;
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Ambiente';
  }

  getSummary() {
    return `T=${this.temp.toFixed(2)}°C, O2=${this.oxygen.toFixed(2)}%, P=${this.pressure.toFixed(2)}hPa, H=${this.humidity.toFixed(2)}%, V=${this.windSpeed.toFixed(2)}m/s, R=${this.radiation.toFixed(2)}µSv ${this.isLastReport ? '✓ Último' : ''}`;
  }
}

// Classe RepairReport
export class RepairReport {
  constructor({ taskType, missionId, problemId, repairable, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.problemId = problemId;
    this.repairable = repairable;
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Reparação';
  }

  getSummary() {
    const status = this.repairable ? '✓ Reparado' : '✗ Não reparável';
    return `Problema #${this.problemId} - ${status} ${this.isLastReport ? '✓ Último' : ''}`;
  }
}

// Classe TopoReport
export class TopoReport {
  constructor({ taskType, missionId, latitude, longitude, height, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.latitude = latitude;
    this.longitude = longitude;
    this.height = height;
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Topografia';
  }

  getSummary() {
    return `(${this.latitude.toFixed(4)}°, ${this.longitude.toFixed(4)}°) h=${this.height.toFixed(2)}m ${this.isLastReport ? '✓ Último' : ''}`;
  }
}

// Classe InstallReport
export class InstallReport {
  constructor({ taskType, missionId, success, isLastReport }) {
    this.taskType = taskType;
    this.missionId = missionId;
    this.success = success;
    this.isLastReport = isLastReport;
  }

  getType() {
    return 'Instalação';
  }

  getSummary() {
    const status = this.success ? '✓ Sucesso' : '✗ Falhou';
    return `${status} ${this.isLastReport ? '✓ Último' : ''}`;
  }
}
