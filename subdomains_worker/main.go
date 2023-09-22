package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/projectdiscovery/subfinder/v2/pkg/runner"
	subfinder "github.com/projectdiscovery/subfinder/v2/pkg/runner"
)

// func processJob(ctx context.Context, args ...interface{}) error {
// 	help := worker.HelperFor(ctx)
// 	log.Printf("Working on job %s, %s\n", help.Jid(), args)

// 	domain_id := args[0].(float64)
// 	domain := args[1].(string)

// 	subfinderOpts := &runner.Options{
// 		Threads:            100, // Thread controls the number of threads to use for active enumerations
// 		Timeout:            30,  // Timeout is the seconds to wait for sources to respond
// 		MaxEnumerationTime: 10,  // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
// 		// ResultCallback: func(s *resolve.HostEntry) {
// 		// callback function executed after each unique subdomain is found
// 		// },
// 		// ProviderConfig: "your_provider_config.yaml",
// 		// and other config related options
// 		RemoveWildcard: true,
// 	}
// 	subfinder, err := runner.NewRunner(subfinderOpts)
// 	if err != nil {
// 		log.Fatalf("failed to create subfinder runner: %v", err)
// 	}
// 	output := &bytes.Buffer{}
// 	if err = subfinder.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
// 		log.Fatalf("failed to enumerate single domain: %v", err)
// 	}
// 	var subdomains []string
// 	for _, s := range strings.Split(strings.Trim(output.String(), " \n"), "\n") {
// 		if len(s) > 0 {
// 			subdomains = append(subdomains, s)
// 		}
// 	}

// 	return help.With(func(cl *faktory.Client) error {
// 		job := faktory.NewJob("SaveSubdomains", domain_id, subdomains)
// 		job.Queue = "ruby_default"
// 		return cl.Push(job)
// 	})
// }

type ParsedSubject struct {
	Topic    string
	Language string
	Priority string
	ClientID string
}

func parseAndValidateSubject(msg jetstream.Msg) (*ParsedSubject, error) {
	subj_parts := strings.Split(msg.Subject(), ".")
	// subject looks like jobs.go.normal.1
	// where go is worker lang, normal is priority (normal|critical), 1 is integer clientID

	ps := &ParsedSubject{
		Topic:    subj_parts[0],
		Language: subj_parts[1],
		Priority: subj_parts[2],
		ClientID: subj_parts[3],
	}

	if ps.Language != "go" {
		return nil, fmt.Errorf("This worker is able to process only jobs.go.* subject")
	}

	return ps, nil
}

func handleMessage(consumeID int, msg jetstream.Msg, runner *subfinder.Runner, nc *nats.Conn) {
	fmt.Printf("Consumer %d recieved message %s\n", consumeID, msg.Subject())

	subj, err := parseAndValidateSubject(msg)
	if err != nil {
		log.Fatal(err)

		if err := msg.Nak(); err != nil {
			log.Fatal("Can't NAK message")
		}
	}

	var payload JobPayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		log.Fatalf("failed to unmarshal job payload: %v", err)
	}
	domain := payload.Params["domain"]

	output := &bytes.Buffer{}

	if err := runner.EnumerateSingleDomainWithCtx(context.Background(), domain, []io.Writer{output}); err != nil {
		log.Fatalf("failed to enumerate single domain: %v", err)
	}
	results := &ResultsPayload{
		Job: "SaveSubdomains",
	}

	for _, s := range strings.Split(strings.Trim(output.String(), " \n"), "\n") {
		if len(s) > 0 {
			results.Params.Domains = append(results.Params.Domains, s)
		}
	}

	res_subj := fmt.Sprintf("results.ruby.%s.%s", subj.Priority, subj.ClientID)

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("failed to marshal results: %v", err)
	}

	if err := nc.Publish(res_subj, resultsJSON); err != nil {
		log.Fatal("Can't publish message")
	}

	msg.Ack()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	nc, err := nats.Connect("nats://local:v3O4Cro25nJUJND1ooL72Jez9tHBYCu0@0.0.0.0:49335")
	if err != nil {
		log.Fatal(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	s, err := js.Stream(ctx, "jobs")
	if err != nil {
		log.Fatal(err)
	}

	cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "ConsumerParallelGo",
		AckPolicy: jetstream.AckExplicitPolicy,
	})

	if err != nil {
		log.Fatal(err)
	}

	subfinderOpts := &runner.Options{
		Threads:            100, // Thread controls the number of threads to use for active enumerations
		Timeout:            30,  // Timeout is the seconds to wait for sources to respond
		MaxEnumerationTime: 10,  // MaxEnumerationTime is the maximum amount of time in mins to wait for enumeration
		// ResultCallback: func(s *resolve.HostEntry) {
		// callback function executed after each unique subdomain is found
		// },
		// ProviderConfig: "your_provider_config.yaml",
		// and other config related options
		// RemoveWildcard: true,
	}
	subfinder, err := runner.NewRunner(subfinderOpts)
	if err != nil {
		log.Fatalf("failed to create subfinder runner: %v", err)
	}

	for i := 0; i < 5; i++ {
		cc, err := cons.Consume(func(consumeID int) jetstream.MessageHandler {
			return func(msg jetstream.Msg) {
				handleMessage(consumeID, msg, subfinder, nc)
			}
		}(i), jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
			fmt.Println(err)
		}))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Consumer %d started\n", i)
		defer cc.Stop()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	// s, err := js.CreateStream(ctx, jetstream.StreamConfig{
	// 	Name:     "TEST_STREAM",
	// 	Subjects: []string{"FOO.*"},
	// })

	// register job types and the function to execute them

}

type JobPayload struct {
	Job    string            `json:"job"`
	Params map[string]string `json:"params"`
}

type ResultsPayload struct {
	Job    string `json:"job"`
	Params struct {
		Domains []string `json:"domains"`
	} `json:"params"`
}
