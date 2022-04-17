package database

import (
	"errors"
	"sync"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoMiscFound = errors.New("no misc found")

type Misc struct {
	gorm.Model
	TermsAndConditions bool
	mu                 sync.Mutex `gorm:"-"`
}

// GetTermsAndConditions returns where the terms and conditions are accepted or not.
func (m *Misc) GetTermsAndConditions() bool {
	return m.TermsAndConditions
}

// CreateTermsAndConditions sets the terms and conditions in the database.
func (m *Misc) CreateTermsAndConditions(accepted bool) error {
	misc := &Misc{
		gorm.Model{
			ID: 1,
		}, accepted, sync.Mutex{},
	}
	if err := saveTermsAndConditions(misc).Error; err != nil {
		return err
	}

	m.TermsAndConditions = accepted
	return nil
}

func FetchMisc() (*Misc, error) {
	var misc Misc
	result := findTermsAndConditions(&misc)
	if err := result.Error; err != nil {
		logging.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to fetch misc from database")
		return nil, err
	}
	return &misc, nil
}
