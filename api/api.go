package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/falun/obsvc/collector"
)

// Handler is a replacement for a raw http.HandlerFunc that ensures all our
// collector API endpoints conform to a consistent request/response format.
type Handler func(*http.Request, map[string]string) Envelope

// HandlerWithStore is a replacement for a raw http.HandlerFunc that injects a
// collector.Store reference.
// TODO: should I just shove this into a context
type HandlerWithStore func(*collector.Store, *http.Request, map[string]string) Envelope

// CollectorHandler is an interface that enables a collection of API endpoints
// to be exposed for a given collector.
type CollectorHandler interface {
	// Id returns the id generated when adding the collector to the store.
	Id() string

	// RegisterEndpoints will be called with the collector.Store that contains
	// the collector as well as two mux.Routers.
	// - baseCollectorRouter enables the collector to attach endpoints to
	//   `<base_path>/collector/` directly if desired. This is a global namespace
	//   and should be used with caution
	// - idCollectorPath enables the collector to attach endpoints to a unique
	//   path for this collector `<base_path>/collector/<id>/`. It can be used-
	//   without concern about conflict with other collectors.
	RegisterEndpoints(store *collector.Store, baseCollectorRouter, idCollectorPath *mux.Router)
}

// Adapt converts an api.Handler into a standard http.HandlerFunc that will
// extract response vars and correctly render an api.Envelope with the correct
// HTTP status code.
func Adapt(h Handler) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		commonApiOutput(h(r, mux.Vars(r)), rw)
	}
}

func AdaptWithStore(c *collector.Store, h HandlerWithStore) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		commonApiOutput(h(c, r, mux.Vars(r)), rw)
	}
}

var NotImplemented = NewError("not_implemented").WithHttpStatus(http.StatusNotImplemented)

func commonApiOutput(envResp Envelope, rw http.ResponseWriter) {
	rw.WriteHeader(envResp.HttpStatus())
	// TODO: encode json or yaml depending on request accept header
	fmt.Fprintf(rw, "%s\n", envResp.EncodeJson(true))
}

// New constructs a new mux.Router that handles api requests to interact with
// the specified collector handlers. API handlers will be namespaced by the
// key in collectorHandlers.
func New(
	store *collector.Store,
	collectorHandlers []CollectorHandler,
) *mux.Router {
	rch := rootCollectorHandler{store}

	baseRouter := mux.NewRouter()
	apiRouter := baseRouter.PathPrefix("/api/").Subrouter()
	collectorRouter := apiRouter.PathPrefix("/collector/").Subrouter()

	apiRouter.HandleFunc("/collectors", Adapt(rch.List))
	apiRouter.HandleFunc("/collector/{id}", Adapt(rch.Get))

	for _, ch := range collectorHandlers {
		collectorPath := fmt.Sprintf("/collector/%v/", ch.Id())
		ch.RegisterEndpoints(store, collectorRouter, apiRouter.PathPrefix(collectorPath).Subrouter())
	}

	baseRouter.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("pong"))

	})
	return baseRouter
}
