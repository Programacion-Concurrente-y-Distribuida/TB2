"""
Pipeline de extraccion, limpieza y augmentation del dataset AQS-EPA.

Curso: Programacion Concurrente y Distribuida - CC65 - UPC
Trabajo Parcial - 2026-01

Etapas:
1. Descarga de 22 archivos annual_conc_by_monitor_YYYY.zip (2004-2025) desde AQS-EPA
2. Carga y consolidacion en un DataFrame
3. Filtrado por los 4 parametros de interes (PM2.5, NO2, O3, CO)
4. Limpieza (3 reglas)
5. Checkpoint: export del DataFrame limpio en CSV
6. Data augmentation (bootstrap + ruido gaussiano) hasta TARGET_ROWS filas
7. Validacion pre/post augmentation
8. Export final en CSV

Fuente: https://aqs.epa.gov/aqsweb/airdata/download_files.html

Uso:
    pip install tqdm requests pandas numpy
    python aqs_pipeline.py
"""

import io
import os
import time
import zipfile
from pathlib import Path
from typing import List

import numpy as np
import pandas as pd
import requests
from tqdm.auto import tqdm

print(f"pandas {pd.__version__} | numpy {np.__version__}")

# ---------------------------------------------------------------------------
# Configuracion
# ---------------------------------------------------------------------------

POLLUTANTS = {
    "88101": "PM2.5",
    "42602": "NO2",
    "44201": "O3",
    "42101": "CO",
}
TARGET_PARAM_CODES = list(POLLUTANTS.keys())

PRIMARY_STANDARDS = {
    "88101": "PM25 Annual 2024",
    "42602": "NO2 Annual 1971",
    "44201": "Ozone 8-hour 2015",
    "42101": "CO 8-hour 1971",
}

KEEP_EVENT_TYPES = {"No Events", "Events Included"}

YEAR_START = 2004
YEAR_END   = 2025
YEARS      = list(range(YEAR_START, YEAR_END + 1))

AQS_URL_TEMPLATE = "https://aqs.epa.gov/aqsweb/airdata/annual_conc_by_monitor_{year}.zip"

TARGET_ROWS       = 1_000_000
NOISE_SIGMA_RATIO = 0.05
RANDOM_SEED       = 42
np.random.seed(RANDOM_SEED)

print(f"Parametros: {list(POLLUTANTS.values())}")
print(f"Anos: {YEAR_START}-{YEAR_END} ({len(YEARS)} archivos)")
print(f"Target final: {TARGET_ROWS:,} filas")

# ---------------------------------------------------------------------------
# Rutas de salida
# ---------------------------------------------------------------------------

BASE_DIR  = Path(__file__).parent / "data"
RAW_DIR   = BASE_DIR / "raw"
CLEAN_CSV = BASE_DIR / "aqs_clean.csv"
FINAL_CSV = BASE_DIR / "aqs_final_3M.csv"

for d in (BASE_DIR, RAW_DIR):
    d.mkdir(parents=True, exist_ok=True)

print(f"Directorio de trabajo: {BASE_DIR}")

# ---------------------------------------------------------------------------
# Descarga
# ---------------------------------------------------------------------------

def download_file(url: str, dest: Path, max_retries: int = 3, timeout: int = 120) -> bool:
    if dest.exists() and dest.stat().st_size > 1024:
        return True

    tmp = dest.with_suffix(dest.suffix + ".part")
    for attempt in range(1, max_retries + 1):
        try:
            with requests.get(url, stream=True, timeout=timeout) as r:
                r.raise_for_status()
                with open(tmp, "wb") as f:
                    for chunk in r.iter_content(chunk_size=1 << 16):
                        f.write(chunk)
            tmp.rename(dest)
            return True
        except requests.RequestException as e:
            print(f"  Intento {attempt}/{max_retries} fallo: {e}")
            if tmp.exists():
                tmp.unlink()
            if attempt < max_retries:
                time.sleep(5 * attempt)
    return False


downloaded, missing = [], []
for year in tqdm(YEARS, desc="Descargando"):
    url  = AQS_URL_TEMPLATE.format(year=year)
    dest = RAW_DIR / f"annual_conc_by_monitor_{year}.zip"
    ok   = download_file(url, dest)
    (downloaded if ok else missing).append(year)

print(f"Descargados/cacheados: {len(downloaded)} archivos")
if missing:
    print(f"NO se pudieron descargar: {missing}")

total_size_mb = sum(p.stat().st_size for p in RAW_DIR.glob("*.zip")) / 1e6
print(f"Tamano total en disco: {total_size_mb:.1f} MB")

# ---------------------------------------------------------------------------
# Carga y filtrado
# ---------------------------------------------------------------------------

USECOLS = [
    "State Code", "County Code", "Site Num",
    "Parameter Code", "POC",
    "Latitude", "Longitude",
    "Parameter Name", "Sample Duration", "Pollutant Standard",
    "Year", "Units of Measure", "Event Type",
    "Observation Count", "Observation Percent", "Completeness Indicator",
    "Valid Day Count", "Required Day Count",
    "Exceptional Data Count", "Null Data Count",
    "Primary Exceedance Count", "Secondary Exceedance Count",
    "Num Obs Below MDL",
    "Arithmetic Mean", "Arithmetic Standard Dev",
    "1st Max Value", "2nd Max Value", "3rd Max Value", "4th Max Value",
    "99th Percentile", "98th Percentile", "95th Percentile",
    "90th Percentile", "75th Percentile", "50th Percentile", "10th Percentile",
    "State Name", "County Name", "City Name", "CBSA Name",
]

DTYPES = {
    "State Code": str, "County Code": str, "Site Num": str, "Parameter Code": str,
}


def load_and_filter_year(year: int) -> pd.DataFrame:
    zip_path = RAW_DIR / f"annual_conc_by_monitor_{year}.zip"
    if not zip_path.exists():
        return pd.DataFrame()

    with zipfile.ZipFile(zip_path) as zf:
        csv_name = next(n for n in zf.namelist() if n.lower().endswith(".csv"))
        with zf.open(csv_name) as fh:
            df = pd.read_csv(fh, usecols=USECOLS, dtype=DTYPES, low_memory=False)

    return df[df["Parameter Code"].isin(TARGET_PARAM_CODES)].copy()


parts = []
for year in tqdm(YEARS, desc="Cargando + filtrando"):
    parts.append(load_and_filter_year(year))

df_raw = pd.concat(parts, ignore_index=True)
del parts
print(f"Filas tras filtro por parametros: {len(df_raw):,}")
print(f"Memoria del DataFrame: {df_raw.memory_usage(deep=True).sum() / 1e6:.1f} MB")

# ---------------------------------------------------------------------------
# Limpieza
# ---------------------------------------------------------------------------

df = df_raw.copy()
print(f"Inicial: {len(df):,} filas")

df = df[df["Completeness Indicator"] == "Y"]
print(f"Tras Completeness=Y: {len(df):,} filas")

df = df[df["Event Type"].isin(KEEP_EVENT_TYPES)]
print(f"Tras Event Type: {len(df):,} filas")

primary_mask = df.apply(
    lambda r: r["Pollutant Standard"] == PRIMARY_STANDARDS.get(r["Parameter Code"]),
    axis=1,
)
df = df[primary_mask]
print(f"Tras Pollutant Standard primario: {len(df):,} filas")

key = ["State Code", "County Code", "Site Num", "POC", "Parameter Code", "Year"]
before = len(df)
df = df.groupby(key, as_index=False).agg({
    **{c: "first" for c in [
        "Latitude", "Longitude", "Parameter Name", "Sample Duration",
        "Pollutant Standard", "Units of Measure", "Event Type",
        "State Name", "County Name", "City Name", "CBSA Name",
    ]},
    **{c: "mean" for c in [
        "Observation Count", "Observation Percent",
        "Valid Day Count", "Required Day Count",
        "Exceptional Data Count", "Null Data Count",
        "Primary Exceedance Count", "Secondary Exceedance Count",
        "Num Obs Below MDL",
        "Arithmetic Mean", "Arithmetic Standard Dev",
        "1st Max Value", "2nd Max Value", "3rd Max Value", "4th Max Value",
        "99th Percentile", "98th Percentile", "95th Percentile",
        "90th Percentile", "75th Percentile", "50th Percentile", "10th Percentile",
    ]},
})
print(f"Tras dedup por monitor-ano-parametro: {len(df):,} filas (antes: {before:,})")

df["pollutant"] = df["Parameter Code"].map(POLLUTANTS)

cobertura = df.groupby(["pollutant", "Year"]).size().unstack(fill_value=0)
print("\nFilas por contaminante x ano:")
print(cobertura)

# ---------------------------------------------------------------------------
# Checkpoint: dataset limpio
# ---------------------------------------------------------------------------

df.to_csv(CLEAN_CSV, index=False)
print(f"\nCheckpoint guardado: {CLEAN_CSV}")
print(f"  {len(df):,} filas x {len(df.columns)} columnas | {CLEAN_CSV.stat().st_size / 1e6:.1f} MB")

# ---------------------------------------------------------------------------
# Data Augmentation
# ---------------------------------------------------------------------------

NUMERIC_COLS = [
    "Observation Count", "Observation Percent",
    "Valid Day Count", "Required Day Count",
    "Exceptional Data Count", "Null Data Count",
    "Primary Exceedance Count", "Secondary Exceedance Count",
    "Num Obs Below MDL",
    "Arithmetic Mean", "Arithmetic Standard Dev",
    "1st Max Value", "2nd Max Value", "3rd Max Value", "4th Max Value",
    "99th Percentile", "98th Percentile", "95th Percentile",
    "90th Percentile", "75th Percentile", "50th Percentile", "10th Percentile",
]

col_stds     = df[NUMERIC_COLS].std(numeric_only=True).fillna(0)
noise_sigmas = (col_stds * NOISE_SIGMA_RATIO).values

print("\nSigma del ruido por columna:")
print(pd.Series(noise_sigmas, index=NUMERIC_COLS).round(4))


def augment_to_target(df_base: pd.DataFrame, target_rows: int) -> pd.DataFrame:
    n_base      = len(df_base)
    n_synthetic = max(0, target_rows - n_base)
    print(f"Filas reales: {n_base:,}  ->  sinteticas a generar: {n_synthetic:,}")

    if n_synthetic == 0:
        out = df_base.copy()
        out["is_synthetic"] = False
        return out

    idx   = np.random.randint(0, n_base, size=n_synthetic)
    synth = df_base.iloc[idx].reset_index(drop=True).copy()

    noise = np.random.normal(loc=0.0, scale=noise_sigmas, size=(n_synthetic, len(NUMERIC_COLS)))
    synth[NUMERIC_COLS] = synth[NUMERIC_COLS].values + noise

    for c in NUMERIC_COLS:
        synth[c] = synth[c].clip(lower=0)

    df_base = df_base.copy()
    df_base["is_synthetic"] = False
    synth["is_synthetic"]   = True

    return pd.concat([df_base, synth], ignore_index=True)


df_final = augment_to_target(df, TARGET_ROWS)
print(f"Tamano dataset final: {len(df_final):,} filas")
print(f"  Reales:     {(~df_final['is_synthetic']).sum():,}")
print(f"  Sinteticas: {df_final['is_synthetic'].sum():,}")

# ---------------------------------------------------------------------------
# Validacion pre/post augmentation
# ---------------------------------------------------------------------------

real  = df_final[~df_final["is_synthetic"]]
synth = df_final[df_final["is_synthetic"]]

summary_cols = ["Arithmetic Mean", "99th Percentile", "50th Percentile", "1st Max Value"]
comp = pd.concat({
    "Real":      real[summary_cols].describe().loc[["mean", "std", "50%"]],
    "Sintetica": synth[summary_cols].describe().loc[["mean", "std", "50%"]],
}, axis=1).round(4)
print("\nComparacion de estadisticas (real vs sintetica):")
print(comp)

print("\nDistribucion por contaminante en dataset final:")
print(df_final.groupby("pollutant").agg(
    n_total=("Year", "size"),
    n_real=("is_synthetic", lambda s: (~s).sum()),
    n_synth=("is_synthetic", "sum"),
))

# ---------------------------------------------------------------------------
# Export final
# ---------------------------------------------------------------------------

df_final.to_csv(FINAL_CSV, index=False)
print(f"\nDataset final guardado: {FINAL_CSV}")
print(f"  {len(df_final):,} filas x {len(df_final.columns)} columnas | {FINAL_CSV.stat().st_size / 1e6:.1f} MB")
