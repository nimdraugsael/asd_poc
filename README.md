PoC of Attack Surface Discovery with:
- NATS
- Rails app
- Ruby workers 
- Go workers

https://www.youtube.com/watch?v=hjXIUPZ7ArM base NATS things
https://www.youtube.com/watch?v=EJJ2SG-cKyM NATS jetstream (durable queue that we use here)

# How to start?

install go, ruby and nats locally
https://nats.io/

start development server (probably you will need to copy and paste nats credentials from your dev server everywhere in code). Sorry, it's hardcoded now
```
nats server run --jetstream 
```
In another tab

restore streams
```
nats stream jobs tmp/nats-js-backup/jobs
nats stream restore tmp/nats-js-backup/results
```

first terminal tab
```
cd rails_app
bundle install
bundle exec rake db:create && bundle exec rake db:migrate 
bundle exec rails c
```

second terminal tab
```
cd subdomains_worker
go mod download
go run main.go
```

third terminal tab. I didn't have enough time to connect it to activerecord, so I only built a forking worker that recieves messages from NATS results stream and prints data to output. We can do it better!
```
cd rails_app
bundle exec WORKERS_COUNT=5 rake save_results_worker:start
```


now you're good to go! let's try in terminal1 (rails console)
```
d = Domain.create!(domain: "example.org")
d.start_subdomain_update!
```
