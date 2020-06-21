package collector

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Collector interface {
	Name() string
	Type() string
	Collect() error
}

type CollectorSnapshot struct {
	Collector         Collector
	LastCollection    time.Time
	LastCollectionErr error
}

//better name?
type Store struct {
	rwMutex             sync.RWMutex
	collectors          map[string]Collector
	lastCollection      map[string]time.Time
	lastCollectionError map[string]error
}

func NewStore() *Store {
	return &Store{
		collectors:          map[string]Collector{},
		lastCollection:      map[string]time.Time{},
		lastCollectionError: map[string]error{},
	}
}

func (store *Store) AddCollector(
	interval time.Duration,
	jitter time.Duration,
	fn Collector,
) (string, context.CancelFunc) {
	collectorId := uuid.Must(uuid.NewV4()).String()
	return store.AddCollectorWithId(collectorId, interval, jitter, fn)
}

func (store *Store) AddCollectorWithId(
	collectorId string,
	interval time.Duration,
	jitter time.Duration,
	fn Collector,
) (string, context.CancelFunc) {
	store.registerCollector(collectorId, fn)

	ctx, cancelFunc := context.WithCancel(context.Background())
	waitTime := applyJitter(interval, jitter)
	log.Printf("Starting collector `%v`, initial collection in `%v`\n", fn.Name(), waitTime)

	timer := time.NewTimer(waitTime)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				when := time.Now()
				err := fn.Collect()
				if err != nil {
					log.Printf("Error collecting `%v`: %v\n", fn.Name(), err)
				}
				store.registerCollection(collectorId, when, err)
			}

			timer.Reset(applyJitter(interval, jitter))
		}
	}()

	return collectorId, cancelFunc
}

func (store *Store) registerCollector(id string, fn Collector) {
	store.rwMutex.Lock()
	defer store.rwMutex.Unlock()

	store.collectors[id] = fn
}

func (store *Store) unregisterCollector(id string) {
	store.rwMutex.Lock()
	defer store.rwMutex.Unlock()

	delete(store.collectors, id)
}

// TODO: the rwMutex here is _way_ to broad since we might have multiple
// collectors reporting at about the same time
func (store *Store) registerCollection(id string, when time.Time, err error) {
	store.rwMutex.Lock()
	defer store.rwMutex.Unlock()

	store.lastCollection[id] = when
	store.lastCollectionError[id] = err
}

// TODO: none of these are thread safe yet; for collector access it's fine
// until we start allowing real-time (de)registration of collectors. For
// Describe it's problematic because of access to the maps

func (store *Store) GetCollectors() map[string]Collector {
	ret := map[string]Collector{}
	for id := range store.collectors {
		ret[id] = store.collectors[id]
	}
	return ret
}

func (store *Store) GetCollectorsOfType(name string) map[string]Collector {
	ret := map[string]Collector{}
	for id, c := range store.collectors {
		if c.Type() == name {
			ret[id] = store.collectors[id]
		}
	}
	return ret
}

func (store *Store) GetCollector(id string) Collector {
	c, ok := store.collectors[id]
	if !ok {
		return nil
	}

	return c
}

func (store *Store) GetSnapshot(id string) *CollectorSnapshot {
	c := store.GetCollector(id)
	if c == nil {
		return nil
	}

	// TODO: whooooooa this is thread unsafe, gonna surprise us some day
	return &CollectorSnapshot{
		Collector:         c,
		LastCollection:    store.lastCollection[id],
		LastCollectionErr: store.lastCollectionError[id],
	}

	/*
		lastCollection := store.lastCollection[id]

		var lastCollectionPtr *string
		if lastCollection.IsZero() {
			lastCollectionPtr = nil
		} else {
			v := fmt.Sprintf("%v", int(lastCollection.UnixNano()/1000000))
			lastCollectionPtr = &v
		}

		err := store.lastCollectionError[id]
		lastCollectionErr := ""
		if err != nil {
			lastCollectionErr = fmt.Sprintf("%v", err)
		}

		return map[string]interface{}{
			"collector_type":        c.Type(),
			"name":                  c.Name(),
			"last_collection_ms":    lastCollectionPtr,
			"last_collection_error": lastCollectionErr,
		}
	*/
}

// applyJitter returns a time.Duration that is base +/- halfJitter
func applyJitter(base time.Duration, halfJitter time.Duration) time.Duration {
	if halfJitter == 0 {
		return base
	}

	offset := rand.Int63n(2*int64(halfJitter)) - int64(halfJitter)
	newBase := int64(base) + offset
	return time.Duration(newBase)
}
