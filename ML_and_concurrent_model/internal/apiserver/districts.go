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
		numValues := map[string]float64{
			"Latitude":  d.Lat,
			"Longitude": d.Lon,
			"Year":      year,
		}
		for k, v := range profile.Numeric {
			numValues[k] = v
		}

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
