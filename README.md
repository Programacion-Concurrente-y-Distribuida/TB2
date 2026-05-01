# Programación Concurrente y Distribuida - TB2

Este repositorio contiene la implementación y el análisis de performance de algoritmos de procesamiento de datos masivos utilizando los paradigmas **Secuencial** y **Concurrente** en Go.

## 📁 Estructura del Proyecto

- `0. Descargar/`: Scripts para la obtención del dataset original.
- `1. Limpieza/`: Unificación y limpieza de archivos CSV.
- `2. Expandir/`: Expansión del dataset original hasta alcanzar **1,000,000 de registros** con ruido sintético.
- `Paradigmas/`: Implementación de los algoritmos de recomendación/procesamiento.
  - `secuencial.go`: Versión monohilo.
  - `concurrente.go`: Versión multihilo optimizada.
  - `helpers.go`: Lógica compartida de procesamiento de texto.
- `benchmark/`: Herramienta de alta precisión para el análisis de métricas.
  - Genera reportes interactivos en `results/charts.html`.
  - Mide Tiempo, CPU (%) y Memoria RAM (MB).

## 🚀 Guía de Ejecución

### 1. Preparación de Datos (1M de registros)
Es necesario generar el dataset antes de ejecutar las pruebas de rendimiento:

```bash
# Unificar y limpiar
cd "1. Limpieza"
go run union_y_limpieza.go

# Expandir a 1 millón
cd "../2. Expandir"
go run expandir_dataset.go
```

### 2. Ejecución del Benchmark
El sistema de benchmark automatiza la compilación, ejecución y generación de gráficos.

```bash
cd benchmark

# Ejecución estándar (100 runs por paradigma)
./benchmark_runner

# Si deseas saltar la compilación de los algoritmos:
./benchmark_runner --skip-compile

# Personalizar el número de ejecuciones:
./benchmark_runner --runs 100 --workers-runs 10
```

## 📊 Análisis de Resultados
Al finalizar el benchmark, se generan los siguientes archivos en `benchmark/results/`:
- **`charts.html`**: Dashboard interactivo con Strip Plots, Boxplots y curvas de escalabilidad.
- **`summary.txt`**: Resumen ejecutivo con el cálculo de Speedup y número óptimo de workers.
- **`exp1_raw.csv`**: Datos crudos de las 200 ejecuciones comparativas.

## 🛠 Requisitos
- **Go 1.22+**
- **macOS/Linux** (para la medición nativa de recursos de sistema)
