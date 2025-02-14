package workers

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
)

// Meta represents a background metadata optimization worker.
type Meta struct {
	conf *config.Config
}

// NewMeta returns a new Meta worker.
func NewMeta(conf *config.Config) *Meta {
	return &Meta{conf: conf}
}

// originalsPath returns the original media files path as string.
func (m *Meta) originalsPath() string {
	return m.conf.OriginalsPath()
}

// Start metadata optimization routine.
func (m *Meta) Start(delay time.Duration) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("metadata: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.MetaWorker.Start(); err != nil {
		return err
	}

	defer mutex.MetaWorker.Stop()

	log.Debugf("metadata: starting routine check")

	settings := m.conf.Settings()
	done := make(map[string]bool)

	limit := 50
	offset := 0
	optimized := 0

	// Run index optimization.
	for {
		photos, err := query.PhotosCheck(limit, offset, delay)

		if err != nil {
			return err
		}

		if len(photos) == 0 {
			break
		} else if offset == 0 {

		}

		for _, photo := range photos {
			if mutex.MetaWorker.Canceled() {
				return errors.New("metadata: check canceled")
			}

			if done[photo.PhotoUID] {
				continue
			}

			done[photo.PhotoUID] = true

			updated, merged, err := photo.Optimize(settings.StackMeta(), settings.StackUUID(), settings.Features.Estimates)

			if err != nil {
				log.Errorf("metadata: %s (optimize photo)", err)
			} else if updated {
				optimized++
				log.Debugf("metadata: optimized photo %s", photo.String())
			}

			for _, p := range merged {
				log.Infof("metadata: merged %s", p.PhotoUID)
				done[p.PhotoUID] = true
			}
		}

		if mutex.MetaWorker.Canceled() {
			return errors.New("metadata: check canceled")
		}

		offset += limit

		time.Sleep(100 * time.Millisecond)
	}

	if optimized > 0 {
		log.Infof("metadata: optimized %d photos", optimized)
	}

	// Explicitly set quality of photos without primary file to -1.
	if err := query.ResetPhotoQuality(); err != nil {
		log.Warnf("metadata: %s (reset quality)", err.Error())
	}

	log.Debugf("metadata: updating photo counts")

	// Update photo counts and visibilities.
	if err := entity.UpdatePhotoCounts(); err != nil {
		log.Warnf("metadata: %s (update counts)", err.Error())
	}

	// Run moments worker.
	if w := photoprism.NewMoments(m.conf); w == nil {
		log.Errorf("metadata: failed updating moments")
	} else if err := w.Start(); err != nil {
		log.Warn(err)
	}

	log.Debugf("metadata: running facial recognition")

	// Run faces worker.
	if w := photoprism.NewFaces(m.conf); w.Disabled() {
		log.Debugf("metadata: skipping facial recognition")
	} else if err := w.Start(photoprism.FacesOptions{}); err != nil {
		log.Warn(err)
	}

	log.Debugf("metadata: updating preview images")

	// Update album, label, and subject preview image hashes.
	if err := query.UpdatePreviews(); err != nil {
		log.Errorf("metadata: %s (update previews)", err)
	}

	// Run garbage collection.
	runtime.GC()

	return nil
}
