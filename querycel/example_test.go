package querycel_test

import (
	"fmt"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

func ExampleCompile() {
	program, err := querycel.Compile("logs", `otel.severityBand(log.severity_number) == "error" && log.body.contains("timeout")`)
	if err != nil {
		panic(err)
	}
	parent := exploration.Query{Signal: "logs", Environment: "production", FromMS: 1000, ToMS: 61000, Limit: 100}
	question, err := program.Apply(parent)
	if err != nil {
		panic(err)
	}
	plan, err := exploration.Compile(question)
	if err != nil {
		panic(err)
	}
	match, err := program.Eval(map[exploration.Field]string{
		{Kind: "column", Key: "severity_number"}: "17",
		{Kind: "column", Key: "body"}:            "upstream timeout",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(querycel.LibraryVersion, question.Environment, match, len(plan.Parameters) > 0)
	// Output: goo.cel.otel/v1 production true true
}
