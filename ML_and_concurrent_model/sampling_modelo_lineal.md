# Prueba del Modelo con una Muestra Individual

## Descripción del muestreo

Para validar el funcionamiento del modelo de regresión lineal ya entrenado, se construyó una muestra individual en formato CSV a partir de un caso externo al conjunto principal del proyecto. La muestra representa un registro de referencia ubicado en Los Ángeles, California, asociado al contaminante `PM2.5` y correspondiente al año 2021. Sus variables geográficas, temporales, observacionales y categóricas mantienen la misma estructura que las empleadas durante el entrenamiento.

La fila se guardó en el archivo `sample_predict.csv`. En esta prueba se incluyó también la columna `Arithmetic Mean` con valor `12.7`, no porque el modelo la necesite para predecir en un escenario real, sino porque el comando actual de inferencia la usa como valor real para calcular el error y mostrar la comparación entre predicción y observación.

## Código usado para la prueba

El archivo de muestreo contiene una cabecera con las columnas requeridas por el modelo y una única observación:

```text
[PLACEHOLDER: insertar captura del archivo sample_predict.csv]
```

Para ejecutar la prueba se utiliza el comando `predict-linear`, indicando la ruta del modelo entrenado y el CSV de entrada:

```bash
cd ML_and_concurrent_model
go run ./cmd/predict-linear -model modelo.json -input sample_predict.csv
```

Si todavía no existe `modelo.json`, primero debe entrenarse y guardarse el modelo:

```bash
cd ML_and_concurrent_model
go run ./cmd/train-linear -input aqs_clean.csv -save-model modelo.json
```

## Resultado obtenido

La ejecución produce una predicción para `Arithmetic Mean` y la compara con el valor real incluido en la muestra (`12.7`). El resultado permite observar tres valores principales: el valor predicho por el modelo, el valor real y el error de predicción.

```text
[PLACEHOLDER: insertar captura de la salida de predict-linear]
```

Resultado registrado:

```text
Predicción obtenida: [COMPLETAR]
Valor real: 12.7
Error: [COMPLETAR]
```

## Capturas requeridas

1. Captura del archivo de muestreo:

```bash
cd ML_and_concurrent_model
cat sample_predict.csv
```

2. Captura del entrenamiento del modelo, solo si necesitas generar `modelo.json`:

```bash
cd ML_and_concurrent_model
go run ./cmd/train-linear -input aqs_clean.csv -save-model modelo.json
```

3. Captura de la inferencia con la muestra:

```bash
cd ML_and_concurrent_model
go run ./cmd/predict-linear -model modelo.json -input sample_predict.csv
```

4. Captura opcional del código del predictor:

```bash
cd ML_and_concurrent_model
sed -n '1,80p' cmd/predict-linear/main.go
```
