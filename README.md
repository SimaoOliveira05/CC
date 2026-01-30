# CC

Sistema de comunicacao para gestao de missoes entre uma nave-mae e rovers em ambiente distribuido.

**Grade:** 20 / 20 ⭐

## Authors

- *Gabriel Dantas* -> [@gabil88](https://github.com/gabil88)
- *José Fernandes* -> [@JoseLourencoFernandes](https://github.com/JoseLourencoFernandes)
- *Simão Oliveira* -> [@SimaoOliveira05](https://github.com/SimaoOliveira05)


## Requisitos

- Go 1.21+
- Node.js 18+ (para o Ground Control)
- npm

## Building

```bash
make build
```

## Executing

### 1. Mothership

```bash
make run-mothership
```

### 2. Rover

In a different machine or terminall from mothership

```bash
make run-rover MS_IP=<MOTHERSHIP-IP>
```

### 3. Ground Control (Dashboard)

```bash
make run-gc MS_IP=<MOTHERSHIP-IP>
```

## Test Mode (with metrics)

```bash
make test-mothership
make test-rover MS_IP=<MOTHERSHIP-IP>
```
