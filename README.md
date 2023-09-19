PoC of Attack Surface Discovery with:
- Faktory
- Rails app
- Ruby workers 
- Go workers

# How to start?

```
docker-compose up -d faktory
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

third terminal tab
```
cd rails_app
bundle exec faktory-worker -q ruby_default,ruby_critical
```

now you're good to go! let's try in terminal1 (rails console)
```
d = Domain.create!(domain: "example.org")
d.start_subdomain_update!
```
