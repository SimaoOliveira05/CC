# 🚀 NasUM Ground Control

Frontend moderno estilo NASA para controle e monitoramento de rovers e missões.

## 🎯 Funcionalidades

- **Dashboard Rovers**: Visualização em tempo real dos rovers ativos com status, bateria e velocidade
- **Gerenciamento de Missões**: Lista de todas as missões ativas com filtros por estado
- **Detalhes de Missões**: Visualização completa dos dados da missão e seus reports
- **Reports Especializados**: Visualização formatada de 6 tipos diferentes de reports:
  - 📸 Captura de Imagem (chunks)
  - 🧪 Coleta de Amostra (componentes químicos)
  - 🌍 Análise Ambiental (temperatura, pressão, etc)
  - 🔧 Reparação/Resgate (status de reparos)
  - 🗺️ Mapeamento Topográfico (coordenadas + Google Maps)
  - ⚙️ Instalação (sucesso/falha)

## 🛠️ Instalação

```bash
cd ground-control
npm install
```

## 🚀 Como Rodar

### Modo Desenvolvimento
```bash
npm run dev
```
Acesse em [http://localhost:5173](http://localhost:5173)

### Build para Produção
```bash
npm run build
npm run preview
```

## ⚙️ Configuração

A API base padrão é `http://localhost:8080/api`. Para mudar, edite em `App.vue`:

```javascript
const API_BASE = 'http://localhost:8080/api'; // altere aqui
```

## 📁 Estrutura do Projeto

```
ground-control/
├── main.js                    # Entry point
├── App.vue                    # App principal (tema NASA)
├── models.js                  # Classes de dados (Rover, Mission, Reports)
├── package.json
├── vite.config.js
├── index.html
└── components/
    ├── RoverCard.vue          # Card individual de rover
    ├── MissionCard.vue        # Card de missão (clicável)
    ├── MissionDetail.vue      # Detalhe completo da missão
    └── reports/               # Componentes de reports
        ├── ImageReportCard.vue
        ├── SampleReportCard.vue
        ├── EnvReportCard.vue
        ├── RepairReportCard.vue
        ├── TopoReportCard.vue
        └── InstallReportCard.vue
```

## 🎨 Tema NASA

O frontend utiliza um tema moderno inspirado em dashboards da NASA com:
- Cores: Azul escuro (#0a1e3d), Cyan (#00d4ff), Verde (#00ff88), Laranja (#ff6b1f)
- Fonte: Courier New (monospace)
- Efeitos: Glow, shadow, animações suaves

## 📊 API Endpoints Esperados

- `GET /api/rovers` - Lista de rovers
- `GET /api/missions` - Lista de missões
- `GET /api/missions/{id}` - Detalhes de uma missão (opcional)

## 💡 Dicas

- Clique em qualquer missão para ver seus reports detalhados
- Os cards de report mostram informações visual e estruturadas
- Cada tipo de report tem cores e ícones distintos
- Os status são coloridos para fácil identificação

## 🔧 Troubleshooting

Se receber erro sobre `.vue` files:
```bash
npm install @vitejs/plugin-vue --save-dev
```

Se a API não conectar, verifique:
1. A URL base em `App.vue`
2. Se a API está rodando em `http://localhost:8080`
3. Se o CORS está habilitado na API

---

Desenvolvido com ❤️ para NasUM 🚀

