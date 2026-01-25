# NasUM - Sistema de Comunicacao Nave-Mae / Rovers

Sistema de comunicacao para gestao de missoes entre uma nave-mae e rovers em ambiente distribuido.

# Grade
**⭐ Grade: ?? / 20 ⭐**

## Authors

- *Gabriel Dantas* -> [@gabil88](https://github.com/gabil88)
- *José Fernandes* -> [@JoseLourencoFernandes](https://github.com/JoseLourencoFernandes)
- *Simão Oliveira* -> [@SimaoOliveira05](https://github.com/SimaoOliveira05)


## Requisitos

- Go 1.21+
- Node.js 18+ (para o Ground Control)
- npm

## Estrutura

```
nasUM/
├── src/              # Backend (Go) - Nave-Mae e Rovers
├── ground-control/   # Frontend (Vue.js) - Dashboard
├── assets/           # Imagens e dados de teste
├── logs/             # Ficheiros de log
└── metrics/          # Metricas de execucao (modo teste)
```

## Compilacao

```bash
make build
```

## Execucao

### 1. Iniciar a Nave-Mae

```bash
make run-mothership
```

### 2. Iniciar um Rover

Numa maquina diferente ou terminal separado:

```bash
make run-rover MS_IP=<IP_DA_NAVE_MAE>
```

Exemplo com nave-mae local:

```bash
make run-rover MS_IP=127.0.0.1
```

### 3. Iniciar o Ground Control (Dashboard)

```bash
make run-gc MS_IP=<IP_DA_NAVE_MAE>
```

Aceder em http://localhost:5173

## Modo de Teste (com metricas)

```bash
make test-mothership
make test-rover MS_IP=<IP_DA_NAVE_MAE>
```

As metricas sao exportadas para ficheiros JSON ao terminar (Ctrl+C).

## Configuracao

O ficheiro `src/config.json` contem os parametros configuraveis:
- Portas de rede
- Timeouts e retransmissoes
- Parametros de bateria e movimento
- Configuracoes dos dispositivos

## Limpeza

```bash
make clean
```

Remove binarios, logs e ficheiros de metricas.
