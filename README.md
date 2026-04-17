# Programacion Concurrente y Distribuida - PC1

Este repositorio contiene los scripts necesarios para unir, limpiar y expandir el dataset de Órdenes de Compra de PERÚ COMPRAS.

## Estructura del Proyecto

- `1. Limpieza/`: Contiene el script `union_y_limpieza.go` para unificar los CSV de la carpeta `/data` y limpiarlos.
- `2. Expandir/`: Contiene el script `expandir_dataset.go` para ampliar el dataset a 1,000,000 de registros con ruido sintético.
- `data/`: Carpeta (ignorada por Git) donde se deben colocar los archivos CSV originales.

## Requisitos

- Go (Golang) instalado.

## Uso

1. Colocar los archivos CSV originales en la carpeta `data/`.
2. Ejecutar la unificación y limpieza:
   ```bash
   cd "1. Limpieza"
   go run union_y_limpieza.go
   ```
3. Ejecutar la expansión a 1M de registros:
   ```bash
   cd "../2. Expandir"
   go run expandir_dataset.go
   ```

El resultado final se encontrará en `data/dataset_1m.csv`.
