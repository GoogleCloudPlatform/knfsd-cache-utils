package main

import (
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
)

func main() {
	factories, err := components()
	if err != nil {
		log.Fatalf("failed to build components: %v", err)
	}

	params := service.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "knfsd-metrics-agent",
			Description: "KNFSD Metrics Agent",
		},
		Factories: factories,
	}

	cmd := service.NewCommand(params)
	err = cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
