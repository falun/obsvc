## TODO List

**Documentation**:
- Write down thoughts about the collector architucture
- How is the API half laid out
- How do collectors add new API endpoints, note this peculiarity
  with the api of collectors such that you may only get one handler for a set
  of like collectors
```
apiHandler := api.New(store, map[string]api.CollectorHandler{
  "foo": api.FooHandler{},
})
```

**General**:
- docs docs docs
- Metrics integration

**Obsvc**
- CORS!
- (or provide an easy awy to route to a webapp sitting next to us out of the box)
- Add aggregators scaffolding
- bake a process for adding custom collectors & aggregators not in the base set
- rewrite API layer to push actual endpoint impl out of the HTTP layer so that
  we can shove alternate transports (grpc) on top if we want

**Collector Framework**:
- handle concurrency safely
- Support configurable lengths of collection history
- Support forced collection on request
- Support collector-based overrides for global attributes, e.g., collection
  interval
- handle concurrency better (at all) _everywhere_
- Build sample collector