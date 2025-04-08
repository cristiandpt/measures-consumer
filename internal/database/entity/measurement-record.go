package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// TelehealthMeasurement represents a document to be stored in MongoDB
// for a single scan from the telehealth device.
type TelehealthMeasurement struct {
	ID                                      primitive.ObjectID `bson:"_id,omitempty"`
	TimeStamp                               string             `bson:"time_stamp,omitempty"`                                  // String
	BloodPressureUnit                       string             `bson:"blood_pressure_unit,omitempty"`                         // String ("mmHg" or "kPa")
	Systolic                                float64            `bson:"systolic,omitempty"`                                    // BigDecimal
	Diastolic                               float64            `bson:"diastolic,omitempty"`                                   // BigDecimal
	MeanArterialPressure                    float64            `bson:"mean_arterial_pressure,omitempty"`                      // BigDecimal
	PulseRate                               float64            `bson:"pulse_rate,omitempty"`                                  // BigDecimal
	BloodPressureMeasurementStatus          []string           `bson:"blood_pressure_measurement_status,omitempty"`           // EnumSet<BloodPressureMeasurementStatus> (represented as array of strings)
	WeightUnit                              string             `bson:"weight_unit,omitempty"`                                 // String ("kg" or "lb")
	HeightUnit                              string             `bson:"height_unit,omitempty"`                                 // String ("m" or "in")
	Weight                                  float64            `bson:"weight,omitempty"`                                      // BigDecimal
	Height                                  float64            `bson:"height,omitempty"`                                      // BigDecimal
	BMI                                     float64            `bson:"bmi,omitempty"`                                         // BigDecimal
	BodyFatPercentage                       float64            `bson:"body_fat_percentage,omitempty"`                         // BigDecimal
	BasalMetabolism                         float64            `bson:"basal_metabolism,omitempty"`                            // BigDecimal (Unit: "kJ")
	MusclePercentage                        float64            `bson:"muscle_percentage,omitempty"`                           // BigDecimal
	MuscleMass                              float64            `bson:"muscle_mass,omitempty"`                                 // BigDecimal (Unit: "kg" or "lb")
	FatFreeMass                             float64            `bson:"fat_free_mass,omitempty"`                               // BigDecimal (Unit: "kg" or "lb")
	SoftLeanMass                            float64            `bson:"soft_lean_mass,omitempty"`                              // BigDecimal (Unit: "kg" or "lb")
	BodyWaterMass                           float64            `bson:"body_water_mass,omitempty"`                             // BigDecimal (Unit: "kg" or "lb")
	SkeletalMusclePercentage                float64            `bson:"skeletal_muscle_percentage,omitempty"`                  // BigDecimal
	VisceralFatLevel                        float64            `bson:"visceral_fat_level,omitempty"`                          // BigDecimal
	BodyAge                                 float64            `bson:"body_age,omitempty"`                                    // BigDecimal
	BodyFatPercentageStageEvaluation        float64            `bson:"body_fat_percentage_stage_evaluation,omitempty"`        // BigDecimal
	SkeletalMusclePercentageStageEvaluation float64            `bson:"skeletal_muscle_percentage_stage_evaluation,omitempty"` // BigDecimal
	VisceralFatLevelStageEvaluation         float64            `bson:"visceral_fat_level_stage_evaluation,omitempty"`         // BigDecimal
	// You might want to add a timestamp for when this record was stored in your system
	CreatedAt time.Time `bson:"created_at,omitempty"`
}
