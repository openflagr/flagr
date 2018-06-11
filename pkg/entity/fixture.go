package entity

import (
	"os"

	"github.com/checkr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // sqlite driver
)

// GenFixtureFlag is a fixture
func GenFixtureFlag() Flag {
	f := Flag{
		ID:          100,
		Description: "",
		Name:        "fixture",
		Enabled:     true,
		Segments:    []Segment{GenFixtureSegment()},
		Variants: []Variant{
			{
				Model:  gorm.Model{ID: 300},
				FlagID: 100,
				Key:    "control",
			},
			{
				Model:  gorm.Model{ID: 301},
				FlagID: 100,
				Key:    "treatment",
				Attachment: map[string]string{
					"value": "321",
				},
			},
		},
	}
	f.PrepareEvaluation()
	return f
}

// GenFixtureSegment is a fixture
func GenFixtureSegment() Segment {
	s := Segment{
		Model:          gorm.Model{ID: 200},
		FlagID:         100,
		Description:    "",
		Rank:           0,
		RolloutPercent: 100,
		Constraints: []Constraint{
			{
				Model:     gorm.Model{ID: 500},
				SegmentID: 200,
				Property:  "dl_state",
				Operator:  models.ConstraintOperatorEQ,
				Value:     `"CA"`,
			},
		},
		Distributions: []Distribution{
			{
				Model:      gorm.Model{ID: 400},
				SegmentID:  200,
				VariantID:  300,
				VariantKey: "control",
				Percent:    50,
			},
			{
				Model:      gorm.Model{ID: 401},
				SegmentID:  200,
				VariantID:  301,
				VariantKey: "treatment",
				Percent:    50,
			},
		},
	}
	s.PrepareEvaluation()
	return s
}

// NewTestDB creates a new in-memory test db
func NewTestDB() *gorm.DB {
	testFile := "/tmp/flagr_test.sqlite"
	os.Remove(testFile)
	db, err := gorm.Open("sqlite3", testFile)
	if err != nil {
		logrus.WithField("err", err).Error("failed to connect to db")
		panic(err)
	}
	db.SetLogger(logrus.StandardLogger())
	db.AutoMigrate(AutoMigrateTables...)
	return db
}

// PopulateTestDB seeds the test db
func PopulateTestDB(flag Flag) *gorm.DB {
	testDB := NewTestDB()
	for _, s := range flag.Segments {
		for _, c := range s.Constraints {
			testDB.Create(&c)
		}
		for _, d := range s.Distributions {
			testDB.Create(&d)
		}
		s.Constraints = []Constraint{}
		s.Distributions = []Distribution{}
		testDB.Create(&s)
	}
	for _, v := range flag.Variants {
		testDB.Create(&v)
	}
	testDB.Create(&flag)

	return testDB
}
