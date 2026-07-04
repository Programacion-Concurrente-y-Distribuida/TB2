package apiserver

import (
	_ "embed"
	"encoding/json"
	"math"
	"net/http"
	"time"

	"aqsml/internal/aqsml"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

//go:embed data/lima_districts.json
var limaDistrictsJSON []byte

// limaDistrict is a district of Lima Metropolitana with a representative
// point (centroid) used as the Latitude/Longitude features fed to the model.
type limaDistrict struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

var limaDistricts = mustLoadLimaDistricts()

func mustLoadLimaDistricts() []limaDistrict {
	var ds []limaDistrict
	if err := json.Unmarshal(limaDistrictsJSON, &ds); err != nil {
		panic("data/lima_districts.json invalido: " + err.Error())
	}
	return ds
}

// districtProfile encodes Lima-specific pollution characteristics per district.
// Values derived from SENAMHI / MINAM monitoring reports (2019-2023).
// obsScale   → relative observation count (activity / traffic proxy, 1.0 = baseline)
// exceedScale → relative exceedance count (how often standards are exceeded)
type districtProfile struct {
	obsScale    float64
	exceedScale float64
}

// limaDistrictProfiles maps district ID to its pollution profile.
// Anchored to real SENAMHI REMCA station data (367 586 hourly PM2.5 records,
// stations: SJL, STA, CRB, VMT, SMP, SBJ, CDM).  Remaining districts
// interpolated from Lima's four geographic zones and land-use characteristics.
// Source: datosabiertos.gob.pe — SENAMHI monitoreo calidad del aire Lima.
var limaDistrictProfiles = map[string]districtProfile{
	// ── Measured by SENAMHI REMCA (real PM2.5 averages) ─────────────────────
	"150132": {1.37, 4.08}, // SJL             — 31.35 µg/m³ (SENAMHI)
	"150137": {1.30, 3.91}, // Santa Anita     — 30.03 µg/m³ (SENAMHI)
	"150106": {1.19, 3.62}, // Carabayllo      — 27.85 µg/m³ (SENAMHI)
	"150143": {1.11, 3.42}, // VMT             — 26.26 µg/m³ (SENAMHI)
	"150135": {0.73, 2.43}, // SMP             — 18.68 µg/m³ (SENAMHI)
	"150130": {0.73, 2.43}, // San Borja       — 18.67 µg/m³ (SENAMHI)
	"150113": {0.65, 2.20}, // Jesús María     — 16.91 µg/m³ (SENAMHI/CDM)

	// ── Lima Norte (interpolado — zona industrial/periurbana) ────────────────
	"150125": {1.15, 3.55}, // Puente Piedra   — conurbado con Carabayllo, ladrilleras
	"150112": {1.10, 3.30}, // Independencia   — eje norte, industria ligera
	"150110": {1.08, 3.20}, // Comas           — denso, cercano a Carabayllo
	"150117": {1.00, 2.90}, // Los Olivos      — comercial-industrial norte
	"150102": {0.55, 1.20}, // Ancón           — costera norte, baja actividad
	"150139": {0.50, 1.00}, // Santa Rosa      — balneario norte

	// ── Lima Este (interpolado — industrial / periurbano) ────────────────────
	"150103": {1.28, 3.80}, // Ate             — zona industrial Huaycan/Vitarte (~STA)
	"150111": {1.20, 3.60}, // El Agustino     — industrial, fábricas (~CRB)
	"150118": {1.05, 3.00}, // Lurigancho      — periurbano, industria ligera
	"150107": {0.68, 1.80}, // Chaclacayo      — residencial, menor densidad

	// ── Lima Centro (interpolado — urbano denso / comercial) ─────────────────
	"150101": {0.90, 2.60}, // Lima Cercado    — centro histórico, alto tráfico
	"150105": {0.85, 2.45}, // Breña           — comercial denso
	"150115": {0.88, 2.50}, // La Victoria     — mercados, tráfico pesado
	"150128": {0.83, 2.35}, // Rímac           — histórico, tráfico elevado
	"150134": {0.78, 2.10}, // San Luis        — industria liviana
	"150116": {0.68, 1.80}, // Lince           — residencial-comercial
	"150141": {0.70, 1.90}, // Surquillo       — comercial-residencial
	"150121": {0.65, 1.70}, // Pueblo Libre    — residencial
	"150131": {0.55, 1.30}, // San Isidro      — financiero, parques (~SBJ)
	"150136": {0.65, 1.70}, // San Miguel      — residencial-costera

	// ── Lima Sur (interpolado — mixto / periurbano) ──────────────────────────
	"150142": {1.18, 3.58}, // Villa El Salv.  — industrial sur, ladrilleras (~VMT)
	"150133": {1.05, 3.05}, // SJM             — periurbano sur, denso
	"150108": {0.72, 1.95}, // Chorrillos      — mixto residencial/costera
	"150140": {0.68, 1.80}, // Santiago Surco  — residencial arbolado
	"150114": {0.58, 1.40}, // La Molina       — residencial, mayor altitud
	"150123": {0.62, 1.55}, // Pachacámac      — periurbano, menor densidad
	"150109": {0.45, 0.80}, // Cieneguilla     — rural, cuenca del Lurín
	"150119": {0.55, 1.20}, // Lurín           — periférico sur
	"150120": {0.62, 1.55}, // Magdalena       — residencial costera

	// ── Balnearios / costera sur (más limpios) ───────────────────────────────
	"150104": {0.52, 1.10}, // Barranco        — residencial, frente al mar
	"150122": {0.50, 1.05}, // Miraflores      — residencial premium, parques
	"150126": {0.38, 0.60}, // Punta Hermosa   — balneario
	"150127": {0.37, 0.55}, // Punta Negra     — balneario
	"150129": {0.37, 0.55}, // San Bartolo     — balneario
	"150138": {0.36, 0.50}, // Sta María Mar   — balneario
	"150124": {0.38, 0.60}, // Pucusana        — costera, muy poca industria
}

// applyDistrictProfile adjusts the base numeric feature map using district-specific
// pollution characteristics. This encodes Lima's real pollution gradient into
// the features that the model uses, producing meaningful spatial variation.
func applyDistrictProfile(base map[string]float64, id string) map[string]float64 {
	prof, ok := limaDistrictProfiles[id]
	if !ok {
		return base
	}
	result := make(map[string]float64, len(base))
	for k, v := range base {
		result[k] = v
	}
	result["Observation Count"] = base["Observation Count"] * prof.obsScale
	result["Primary Exceedance Count"] = base["Primary Exceedance Count"] * prof.exceedScale
	result["Secondary Exceedance Count"] = base["Secondary Exceedance Count"] * prof.exceedScale
	return result
}

// breakpoint is one step of an EPA-style AQI scale: values <= Max fall in Level.
type breakpoint struct {
	Max   float64
	Level string
	Color string
}

// pollutantProfile fija los valores categoricos y numericos "tipicos" (medias
// del propio dataset AQS de entrenamiento) para un contaminante. El modelo se
// entrena con estaciones de monitoreo de EE.UU., que no comparten columnas
// como "Observation Count" con estaciones peruanas, asi que usamos el promedio
// historico de esa columna para ese contaminante como valor por defecto,
// dejando que solo Latitude/Longitude/Year varien por distrito.
type pollutantProfile struct {
	Unit              string
	ParameterCode     string
	SampleDuration    string
	PollutantStandard string
	UnitsOfMeasure    string
	EventType         string
	Numeric           map[string]float64
	Breakpoints       []breakpoint
}

var pollutantProfiles = map[string]pollutantProfile{
	"O3": {
		Unit:              "ppm",
		ParameterCode:     "44201",
		SampleDuration:    "8-HR RUN AVG BEGIN HOUR",
		PollutantStandard: "Ozone 8-hour 2015",
		UnitsOfMeasure:    "Parts per million",
		EventType:         "No Events",
		Numeric: map[string]float64{
			"Observation Count":          5026.30,
			"Observation Percent":        96.36,
			"Valid Day Count":            254.69,
			"Required Day Count":         264.61,
			"Exceptional Data Count":     0,
			"Null Data Count":            0,
			"Num Obs Below MDL":          0,
			"Primary Exceedance Count":   7.16,
			"Secondary Exceedance Count": 7.16,
		},
		Breakpoints: []breakpoint{
			{0.054, "Bueno", "#00e400"},
			{0.070, "Moderado", "#ffff00"},
			{0.085, "Dañino (grupos sensibles)", "#ff7e00"},
			{0.105, "Dañino", "#ff0000"},
			{0.200, "Muy dañino", "#8f3f97"},
			{math.MaxFloat64, "Peligroso", "#7e0023"},
		},
	},
	"PM2.5": {
		Unit:              "µg/m³",
		ParameterCode:     "88101",
		SampleDuration:    "24 HOUR",
		PollutantStandard: "PM25 Annual 2024",
		UnitsOfMeasure:    "Micrograms/cubic meter (LC)",
		EventType:         "No Events",
		Numeric: map[string]float64{
			"Observation Count":          206.58,
			"Observation Percent":        95.81,
			"Valid Day Count":            204.51,
			"Required Day Count":         212.97,
			"Exceptional Data Count":     1.06,
			"Null Data Count":            4.12,
			"Num Obs Below MDL":          0,
			"Primary Exceedance Count":   39.94,
			"Secondary Exceedance Count": 13.62,
		},
		Breakpoints: []breakpoint{
			{12.0, "Bueno", "#00e400"},
			{35.4, "Moderado", "#ffff00"},
			{55.4, "Dañino (grupos sensibles)", "#ff7e00"},
			{150.4, "Dañino", "#ff0000"},
			{250.4, "Muy dañino", "#8f3f97"},
			{math.MaxFloat64, "Peligroso", "#7e0023"},
		},
	},
	"NO2": {
		Unit:              "ppb",
		ParameterCode:     "42602",
		SampleDuration:    "1 HOUR",
		PollutantStandard: "NO2 Annual 1971",
		UnitsOfMeasure:    "Parts per billion",
		EventType:         "No Events",
		Numeric: map[string]float64{
			"Observation Count":          8181.39,
			"Observation Percent":        93.31,
			"Valid Day Count":            347.75,
			"Required Day Count":         365.30,
			"Exceptional Data Count":     121.27,
			"Null Data Count":            515.48,
			"Num Obs Below MDL":          0,
			"Primary Exceedance Count":   0,
			"Secondary Exceedance Count": 0,
		},
		Breakpoints: []breakpoint{
			{53, "Bueno", "#00e400"},
			{100, "Moderado", "#ffff00"},
			{360, "Dañino (grupos sensibles)", "#ff7e00"},
			{649, "Dañino", "#ff0000"},
			{1249, "Muy dañino", "#8f3f97"},
			{math.MaxFloat64, "Peligroso", "#7e0023"},
		},
	},
	"CO": {
		Unit:              "ppm",
		ParameterCode:     "42101",
		SampleDuration:    "8-HR RUN AVG END HOUR",
		PollutantStandard: "CO 8-hour 1971",
		UnitsOfMeasure:    "Parts per million",
		EventType:         "No Events",
		Numeric: map[string]float64{
			"Observation Count":          7623.44,
			"Observation Percent":        86.98,
			"Valid Day Count":            324.00,
			"Required Day Count":         364.85,
			"Exceptional Data Count":     0,
			"Null Data Count":            0,
			"Num Obs Below MDL":          0,
			"Primary Exceedance Count":   0.24,
			"Secondary Exceedance Count": 0.24,
		},
		Breakpoints: []breakpoint{
			{4.4, "Bueno", "#00e400"},
			{9.4, "Moderado", "#ffff00"},
			{12.4, "Dañino (grupos sensibles)", "#ff7e00"},
			{15.4, "Dañino", "#ff0000"},
			{30.4, "Muy dañino", "#8f3f97"},
			{math.MaxFloat64, "Peligroso", "#7e0023"},
		},
	},
}

func classify(bps []breakpoint, value float64) (level, color string) {
	for _, bp := range bps {
		if value <= bp.Max {
			return bp.Level, bp.Color
		}
	}
	last := bps[len(bps)-1]
	return last.Level, last.Color
}

// ---- GET /api/districts ----

func (s *Server) handleListDistricts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, limaDistricts)
}

// ---- GET /api/districts/predict?pollutant=PM2.5 ----

type districtPrediction struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Prediction float64 `json:"prediction"`
	Unit       string  `json:"unit"`
	Level      string  `json:"level"`
	Color      string  `json:"color"`
}

func (s *Server) handleDistrictPredictions(w http.ResponseWriter, r *http.Request) {
	pollutant := r.URL.Query().Get("pollutant")
	if pollutant == "" {
		pollutant = "PM2.5"
	}
	profile, ok := pollutantProfiles[pollutant]
	if !ok {
		writeError(w, http.StatusBadRequest, "contaminante no soportado: "+pollutant)
		return
	}

	model, err := s.store.LatestModel(r.Context())
	if err == mongo.ErrNoDocuments {
		writeError(w, http.StatusNotFound, "ningún modelo entrenado disponible todavía")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	year := float64(time.Now().Year())
	catValues := map[string]string{
		"pollutant":          pollutant,
		"Parameter Code":     profile.ParameterCode,
		"Sample Duration":    profile.SampleDuration,
		"Event Type":         profile.EventType,
		"Units of Measure":   profile.UnitsOfMeasure,
		"Pollutant Standard": profile.PollutantStandard,
		// "State Code" se omite a proposito: Lima no tiene codigo de estado de
		// EE.UU., asi que queda sin match en el encoding y no aporta senal
		// (equivale al nivel base de esa categoria).
	}

	results := make([]districtPrediction, 0, len(limaDistricts))
	for _, d := range limaDistricts {
		baseNum := map[string]float64{
			"Latitude":  d.Lat,
			"Longitude": d.Lon,
			"Year":      year,
		}
		for k, v := range profile.Numeric {
			baseNum[k] = v
		}
		numValues := applyDistrictProfile(baseNum, d.ID)

		x, err := aqsml.EncodeInputJSON(model.NumFeatureNames, model.PlanJSON, numValues, catValues, model.Means, model.Stds)
		if err != nil {
			continue
		}
		var pred float64
		for i, b := range model.Beta {
			if i < len(x) {
				pred += b * x[i]
			}
		}
		if pred < 0 {
			pred = 0
		}
		level, color := classify(profile.Breakpoints, pred)
		results = append(results, districtPrediction{
			ID: d.ID, Name: d.Name, Lat: d.Lat, Lon: d.Lon,
			Prediction: pred, Unit: profile.Unit, Level: level, Color: color,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pollutant":   pollutant,
		"unit":        profile.Unit,
		"modelId":     model.ID.Hex(),
		"trainedAt":   model.TrainedAt,
		"generatedAt": time.Now(),
		"districts":   results,
	})
}
