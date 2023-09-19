package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"strings"

	faktory "github.com/contribsys/faktory/client"
	worker "github.com/contribsys/faktory_worker_go"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

func processJob(ctx context.Context, args ...interface{}) error {
	help := worker.HelperFor(ctx)
	log.Printf("Working on job %s, %s\n", help.Jid(), args)

	domain_id := args[0].(float64)
	domain := args[1].(string)

	subfinderOpts := &runner.Options{
		Threads:            100, // Thread controls the number of threads to use for active enumerations
		Timeout:            30,  // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: 10,  // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		// ResultCallback: func(s *resolve.HostEntry) {
		// callback function executed after each unique subdomain is found
		// },
		// ProviderConfig: "your_provider_config.yaml",
		// and other config related options
		RemoveWildcard: true,
	}
	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		log.Fatalf("failed to create subfinder runner: %v", err)
	}
	output := &bytes.Buffer{}
	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
		log.Fatalf("failed to enumerate single domain: %v", err)
	}
	var subdomains []string
	for _, s := range strings.Split(strings.Trim(output.String(), " \n"), "\n") {
		if len(s) > 0 {
			subdomains = append(subdomains, s)
		}
	}

	return help.With(func(cl *faktory.Client) error {
		job := faktory.NewJob("SaveSubdomains", domain_id, subdomains)
		job.Queue = "ruby_default"
		return cl.Push(job)
	})
}

func main() {
	mgr := worker.NewManager()

	// register job types and the function to execute them
	mgr.Register("EnumerateSubdomains", processJob)
	//mgr.Register("AnotherJob", anotherFunc)

	// use up to N goroutines to execute jobs
	mgr.Concurrency = 20

	// pull jobs from these queues, in this order of precedence
	mgr.ProcessStrictPriorityQueues("go_critical", "go_default")

	// alternatively you can use weights to avoid starvation
	//mgr.ProcessWeightedPriorityQueues(map[string]int{"critical":3, "default":2, "bulk":1})

	// Start processing jobs, this method does not return.
	mgr.Run()
}
