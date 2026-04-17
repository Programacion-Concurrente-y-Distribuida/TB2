import os
import urllib.request
import time

def main():
    # El archivo base a buscar tiene la estructura:
    # https://www.datosabiertos.gob.pe/sites/default/files/ReportePCBienesYYYYMM.csv
    base_url = "https://www.datosabiertos.gob.pe/sites/default/files/ReportePCBienes"
    
    os.makedirs('data', exist_ok=True)
    
    # Queremos descargar desde 2022-01 hasta 2026-03
    years = [2022, 2023, 2024, 2025, 2026]
    
    print("[*] Iniciando descarga directa de los reportes mensuales de Órdenes de Compra...")
    
    total_intentos = 0
    descargados = 0
    
    for year in years:
        for month in range(1, 13):
            # Si estamos en 2026, detenernos en el mes 03 según indicaste
            if year == 2026 and month > 3:
                break
                
            mes_texto = f"{month:02d}" # Formatea el mes a dos dígitos, ej: "08"
            filename = f"ReportePCBienes{year}{mes_texto}.csv"
            url = f"{base_url}{year}{mes_texto}.csv"
            filepath = os.path.join('data', filename)
            
            total_intentos += 1
            print(f"[*] Buscando archivo: {filename}")
            
            try:
                # Verificamos si existe el archivo para evitar descargar el mismo error 404
                req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
                response = urllib.request.urlopen(req)
                
                # Si llegamos aquí es porque existe
                with open(filepath, 'wb') as f:
                    f.write(response.read())
                    
                print(f"  -> Descargado con exito: {filename}")
                descargados += 1
                time.sleep(0.5) # Pausa amigable para el servidor
                
            except urllib.error.HTTPError as e:
                if e.code == 404:
                    print(f"  -> No encontrado: El archivo {filename} no esta publicado.")
                else:
                    print(f"  -> Error HTTP al descargar {filename}: {e.code}")
            except Exception as e:
                print(f"  -> Error al descargar {url}\n     Detalle: {e}")

    print(f"\n[Terminado] Se han descargado correctamente {descargados} de {total_intentos} intentos. Revise la carpeta 'data'.")

if __name__ == "__main__":
    main()
