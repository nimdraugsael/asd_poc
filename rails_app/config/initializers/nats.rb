require 'nats/io/client'

# Establish connection to NATS

opts = {
  servers: [
    "nats://local:v3O4Cro25nJUJND1ooL72Jez9tHBYCu0@0.0.0.0:49335"
  ],
  reconnect_time_wait: 0.5,
  max_reconnect_attempts: 5
}
$nats = NATS::IO::Client.new
$nats = NATS.connect(opts)

# $nats = NATS::IO::Client.new
# $nats.connect(servers: ["nats://127.0.0.1:4222"])

$nats.on_error do |e|
  Rails.logger.warn "NATS - Error: #{e}"
end

$nats.on_reconnect do
  Rails.logger.warn "NATS - Reconnected!"
end
