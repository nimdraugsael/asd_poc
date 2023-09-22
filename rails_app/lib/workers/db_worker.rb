require File.expand_path('../../../config/environment', __FILE__)
require 'nats/io/client'
require 'fileutils'
require 'thread'

FileUtils.mkdir_p('./tmp/db_workers')
File.write('./tmp/db_workers/master.pid', Process.pid)

opts = {
  servers: [
    ENV.fetch('NATS_URI', "nats://local:v3O4Cro25nJUJND1ooL72Jez9tHBYCu0@0.0.0.0:49335")
  ],
  reconnect_time_wait: 0.5,
  max_reconnect_attempts: 5
}

puts "NATS options #{opts}"

worker_pids = []
shutdown_queue = Queue.new
worker_count = ENV.fetch('WORKERS_COUNT', 3).to_i
worker_pids = []

Signal.trap('TERM') do
  puts "\nCaught TERM, exiting..."
  worker_count.times { shutdown_queue << true }
end

Signal.trap('INT') do
  puts "\nCaught Ctrl+C, exiting..."
  worker_count.times { shutdown_queue << true }
end

Process.setproctitle("dbworker - master")
puts "Started master process: #{Process.pid}"
puts "Starting #{worker_count} worker processes"

worker_count.times do |i|
  worker_pid = fork do
    Process.setproctitle("dbworker - worker")

    nats = NATS::IO::Client.new
    nats.connect(opts)
    psub = nats.jetstream.pull_subscribe("results.*.*.*", "results")

    puts("Started worker process: #{Process.pid}")

    loop do
      if !shutdown_queue.empty?
        puts "Shutting down worker with pid #{Process.pid}"
        break
      end

      begin
        msgs = psub.fetch(1)
        msgs.each do |msg|
          puts("Message got #{msg.subject}")
          data = JSON.parse(msg.data)
          domain_id = data['domain_id']
          subdomains = data['subdomains']
          puts("Got message", data)
          msg.ack
        end
      rescue NATS::IO::Timeout
        # pass
      end
    end
  end

  worker_pids << worker_pid
end

worker_pids.each { |pid| Process.waitpid(pid) }
FileUtils.rm_f('./tmp/db_workers/master.pid')
