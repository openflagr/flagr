package handler

import (
	"sync"
	"time"

	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/pkg/datar"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

var (
	singletonDatar     *Datar
	singletonDatarOnce sync.Once
)

// Datar is the aggregate analytics engine.
// It stores evaluation counts in-memory and periodically flushes them to the DB.
type Datar struct {
	aggregator    *datar.Aggregator
	store         *datar.Store
	flushInterval time.Duration
	closeCh       chan struct{}
	wg            sync.WaitGroup
}

// GetDatar returns the singleton Datar instance.
// Creates the instance on first call, starting its flush loop.
// Returns nil if Datar is not enabled.
func GetDatar() *Datar {
	singletonDatarOnce.Do(func() {
		if !config.Config.DatarEnabled {
			return
		}
		singletonDatar = &Datar{
			aggregator: datar.NewAggregator(),
			store:      datar.NewTestStore(getDB()),

			flushInterval: config.Config.DatarFlushInterval,
			closeCh:       make(chan struct{}),
		}
		singletonDatar.start()
		logrus.Info("Datar: started aggregate analytics")
	})
	return singletonDatar
}

// Record logs an evaluation result into the in-memory aggregator.
// Safe to call from concurrent goroutines.
func (d *Datar) Record(r *models.EvalResult) {
	if d == nil {
		return
	}
	d.aggregator.Record(r)
}

// Shutdown flushes the in-memory buffer to DB and stops the flush loop.
func (d *Datar) Shutdown() error {
	if d == nil {
		return nil
	}

	logrus.Info("Datar: shutting down")

	d.aggregator.Close()
	close(d.closeCh)
	d.wg.Wait()

	agg := d.aggregator.SnapshotAndReset()
	if len(agg) > 0 {
		logrus.WithField("keys", len(agg)).Info("Datar: flushing remaining aggregates on shutdown")
		if err := d.store.FlushAggregates(agg); err != nil {
			logrus.WithError(err).Error("Datar: shutdown flush failed, data may be lost")
			return err
		}
	}

	logrus.Info("Datar: shutdown complete")
	return nil
}

// Len returns the number of keys in the buffer (for health check).
func (d *Datar) Len() int {
	return d.aggregator.Len()
}

// FlushInterval returns the configured flush interval (for health check).
func (d *Datar) FlushInterval() time.Duration {
	return d.flushInterval
}

// Store returns the underlying store (for HTTP handlers).
func (d *Datar) Store() *datar.Store {
	return d.store
}

func (d *Datar) start() {
	d.wg.Add(1)
	go d.flushLoop()
}

func (d *Datar) flushLoop() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.closeCh:
			return
		case <-ticker.C:
			d.flush()
		}
	}
}

func (d *Datar) flush() {
	if d.aggregator.Len() == 0 {
		return
	}

	agg := d.aggregator.SnapshotAndReset()
	if len(agg) == 0 {
		return
	}

	logrus.WithField("keys", len(agg)).Debug("Datar: flushing aggregates")

	if err := d.store.FlushAggregates(agg); err != nil {
		logrus.WithError(err).Error("Datar: flush failed, data in this cycle may be lost")
	}
}

// init ensures GetDatar is called during startup if DatarEnabled.
// It's used by the Setup function to register handlers and shutdown hooks.
func init() {
	// No-op init hook — GetDatar is called explicitly from Setup.
}
