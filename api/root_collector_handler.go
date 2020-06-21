package api

import (
	"net/http"

	"github.com/falun/obsvc/collector"
	"github.com/falun/obsvc/util"
)

type rootCollectorHandler struct {
	store *collector.Store
}

// List returns a map of collector type to a list of ids of registered
// collectors
func (rch *rootCollectorHandler) List(
	r *http.Request,
	_ map[string]string,
) Envelope {
	cs := rch.store.GetCollectors()

	ret := map[string][]string{}

	// walk the collectors
	for cId, ch := range cs {
		ctype := ch.Type()
		if _, ok := ret[ctype]; !ok {
			ret[ctype] = []string{}
		}
		// for the collector type collect an id associated with it
		ret[ctype] = append(ret[ctype], cId)
	}

	return NewResponse(ret)
}

// Get returns an Envelope containing a description of the collector indicated
// by a variable named "id".
func (rch *rootCollectorHandler) Get(
	r *http.Request,
	vars map[string]string,
) Envelope {
	collectorId := vars["id"]

	ssData := rch.store.GetSnapshot(collectorId)
	if collectorId == "" || ssData == nil {
		return NewError("unknown_collector").WithHttpStatus(http.StatusNotFound)
	}

	return NewResponse(
		map[string]interface{}{
			"name":                  ssData.Collector.Name(),
			"type":                  ssData.Collector.Type(),
			"last_collection_ms":    util.TimeToMS(ssData.LastCollection),
			"last_collection_error": util.ErrToString(ssData.LastCollectionErr),
		},
	)
}
