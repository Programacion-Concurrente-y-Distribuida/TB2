package aqsml

import (
	"context"
	"fmt"
)

// MeasurementCursor is the minimal interface needed to stream documents from
// the measurements collection without importing mongo-driver into this package.
type MeasurementCursor interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	Err() error
	Close(ctx context.Context) error
}

// MeasurementDoc mirrors the fields stored by migrate-mongo that are needed
// to build a rowData. Only the fields used as features + target are decoded.
type MeasurementDoc struct {
	Year                  int     `bson:"Year"`
	Latitude              float64 `bson:"Latitude"`
	Longitude             float64 `bson:"Longitude"`
	ObservationCount      float64 `bson:"Observation Count"`
	ObservationPercent    float64 `bson:"Observation Percent"`
	ValidDayCount         float64 `bson:"Valid Day Count"`
	RequiredDayCount      float64 `bson:"Required Day Count"`
	ExceptionalDataCount  float64 `bson:"Exceptional Data Count"`
	NullDataCount         float64 `bson:"Null Data Count"`
	NumObsBelowMDL        float64 `bson:"Num Obs Below MDL"`
	PrimaryExceedance     float64 `bson:"Primary Exceedance Count"`
	SecondaryExceedance   float64 `bson:"Secondary Exceedance Count"`
	ArithmeticMean        float64 `bson:"Arithmetic Mean"`
	Pollutant             string  `bson:"pollutant"`
	ParameterCode         string  `bson:"Parameter Code"`
	SampleDuration        string  `bson:"Sample Duration"`
	EventType             string  `bson:"Event Type"`
	UnitsOfMeasure        string  `bson:"Units of Measure"`
	PollutantStandard     string  `bson:"Pollutant Standard"`
	StateCode             string  `bson:"State Code"`
}

// LoadRowsFromMongo streams all documents from the cursor and converts them
// to the []rowData format used by the rest of the ML pipeline.
// maxRows == 0 means no limit.
func LoadRowsFromMongo(ctx context.Context, cursor MeasurementCursor, maxRows int) ([]rowData, error) {
	defer cursor.Close(ctx)

	out := make([]rowData, 0, 65536)
	var skipped int

	for cursor.Next(ctx) {
		var doc MeasurementDoc
		if err := cursor.Decode(&doc); err != nil {
			skipped++
			continue
		}

		if doc.Year == 0 || doc.ArithmeticMean == 0 {
			skipped++
			continue
		}
		if doc.Pollutant == "" || doc.ParameterCode == "" {
			skipped++
			continue
		}

		row := rowData{
			Year: doc.Year,
			NumValues: []float64{
				doc.Latitude,
				doc.Longitude,
				float64(doc.Year),
				doc.ObservationCount,
				doc.ObservationPercent,
				doc.ValidDayCount,
				doc.RequiredDayCount,
				doc.ExceptionalDataCount,
				doc.NullDataCount,
				doc.NumObsBelowMDL,
				doc.PrimaryExceedance,
				doc.SecondaryExceedance,
			},
			CatValues: []string{
				doc.Pollutant,
				doc.ParameterCode,
				doc.SampleDuration,
				doc.EventType,
				doc.UnitsOfMeasure,
				doc.PollutantStandard,
				doc.StateCode,
			},
			Target: doc.ArithmeticMean,
		}

		out = append(out, row)

		if maxRows > 0 && len(out) >= maxRows {
			break
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no se encontraron documentos válidos en la colección measurements (skipped=%d)", skipped)
	}

	return out, nil
}
