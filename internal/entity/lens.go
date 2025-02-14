package entity

import (
	"strings"
	"sync"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

var lensMutex = sync.Mutex{}

// Lenses represents a list of lenses.
type Lenses []Lens

// Lens represents camera lens (as extracted from UpdateExif metadata)
type Lens struct {
	ID              uint       `gorm:"primary_key" json:"ID" yaml:"ID"`
	LensSlug        string     `gorm:"type:VARBINARY(255);unique_index;" json:"Slug" yaml:"Slug,omitempty"`
	LensName        string     `gorm:"type:VARCHAR(255);" json:"Name" yaml:"Name"`
	LensMake        string     `gorm:"type:VARCHAR(255);" json:"Make" yaml:"Make,omitempty"`
	LensModel       string     `gorm:"type:VARCHAR(255);" json:"Model" yaml:"Model,omitempty"`
	LensType        string     `gorm:"type:VARCHAR(255);" json:"Type" yaml:"Type,omitempty"`
	LensDescription string     `gorm:"type:TEXT;" json:"Description,omitempty" yaml:"Description,omitempty"`
	LensNotes       string     `gorm:"type:TEXT;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt       time.Time  `json:"-" yaml:"-"`
	UpdatedAt       time.Time  `json:"-" yaml:"-"`
	DeletedAt       *time.Time `sql:"index" json:"-" yaml:"-"`
}

var UnknownLens = Lens{
	LensSlug:  UnknownID,
	LensName:  "Unknown",
	LensMake:  "",
	LensModel: "Unknown",
}

// CreateUnknownLens initializes the database with an unknown lens if not exists
func CreateUnknownLens() {
	FirstOrCreateLens(&UnknownLens)
}

// TableName returns the entity database table name.
func (Lens) TableName() string {
	return "lenses"
}

// NewLens creates a new lens in database
func NewLens(modelName string, makeName string) *Lens {
	modelName = txt.Clip(modelName, txt.ClipDefault)
	makeName = txt.Clip(makeName, txt.ClipDefault)

	if modelName == "" && makeName == "" {
		return &UnknownLens
	} else if strings.HasPrefix(modelName, makeName) {
		modelName = strings.TrimSpace(modelName[len(makeName):])
	}

	if n, ok := CameraMakes[makeName]; ok {
		makeName = n
	}

	var name []string

	if makeName != "" {
		name = append(name, makeName)
	}

	if modelName != "" {
		name = append(name, modelName)
	}

	lensName := strings.Join(name, " ")
	lensSlug := slug.Make(txt.Clip(lensName, txt.ClipSlug))

	result := &Lens{
		LensSlug:  lensSlug,
		LensName:  lensName,
		LensMake:  makeName,
		LensModel: modelName,
	}

	return result
}

// Create inserts a new row to the database.
func (m *Lens) Create() error {
	lensMutex.Lock()
	defer lensMutex.Unlock()

	return Db().Create(m).Error
}

// FirstOrCreateLens returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateLens(m *Lens) *Lens {
	if m.LensSlug == "" {
		return &UnknownLens
	}

	if cacheData, ok := lensCache.Get(m.LensSlug); ok {
		log.Debugf("lens: cache hit for %s", m.LensSlug)

		return cacheData.(*Lens)
	}

	result := Lens{}

	if res := Db().Where("lens_slug = ?", m.LensSlug).First(&result); res.Error == nil {
		lensCache.SetDefault(m.LensSlug, &result)
		return &result
	} else if err := m.Create(); err == nil {
		if !m.Unknown() {
			event.EntitiesCreated("lenses", []*Lens{m})

			event.Publish("count.lenses", event.Data{
				"count": 1,
			})
		}

		lensCache.SetDefault(m.LensSlug, m)

		return m
	} else if res := Db().Where("lens_slug = ?", m.LensSlug).First(&result); res.Error == nil {
		lensCache.SetDefault(m.LensSlug, &result)
		return &result
	} else {
		log.Errorf("lens: %s (create %s)", err.Error(), txt.Quote(m.String()))
	}

	return &UnknownLens
}

// String returns an identifier that can be used in logs.
func (m *Lens) String() string {
	return m.LensName
}

// Unknown returns true if the lens is not a known make or model.
func (m *Lens) Unknown() bool {
	return m.LensSlug == "" || m.LensSlug == UnknownLens.LensSlug
}
