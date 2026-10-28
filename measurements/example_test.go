package measurements_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/casselc/goo/measurements"
)

// Supplied rows use the existing native source-adapter representation. Evaluation
// needs neither a ClickHouse connection nor a publication/custody directory.
// This service-bound profile requires the reported service.instance.id.
func ExampleEvaluate() {
	definition, err := measurements.Read(`
apiVersion: goo/measurement-package/v1alpha1
name: queue-depth
title: Pending work
kind: gauge
environment: demo
lookbackMinutes: 1
maxGapMs: 30000
grain: per-stream
source:
  services: [worker]
  scope: example.worker
  version: "1"
  schemaURL: https://opentelemetry.io/schemas/1.37.0
  metric: app.queue.pending
  unit: '{item}'
`)
	if err != nil {
		panic(err)
	}
	question := measurements.Question{Definition: definition, FromMS: 200000, ToMS: 220000}
	rows := []json.RawMessage{json.RawMessage(`{
  "resource":{"service.name":"worker","service.instance.id":"worker-1","deployment.environment.name":"demo"},
  "resource_schema_url":"","scope":"example.worker","version":"1",
  "schema_url":"https://opentelemetry.io/schemas/1.37.0",
  "name":"app.queue.pending","unit":"{item}","attributes":{},
  "end_ns":"210000000000","value":"7","flags":0
 }`)}
	result, err := measurements.Evaluate(question, rows, time.UnixMilli(220000))
	if err != nil {
		panic(err)
	}
	series := result.Gauges.Series[0]
	fmt.Println(result.State, result.Count, series.Stream.Unit, *series.Samples[0].Value)
	// Output: observed 1 {item} 7
}
