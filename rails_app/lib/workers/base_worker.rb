require File.expand_path('../../../config/environment', __FILE__)
require 'nats/io/client'
require 'thread'

worker_count = 5
puts "Started master process: #{Process.pid}"
puts "Starting #{worker_count} worker processes"
Process.setproctitle("fooworker: master")

worker_pids = []
shutdown_queue = Queue.new

Signal.trap('INT') do
  puts "\nCaught Ctrl+C, exiting..."
  worker_count.times { shutdown_queue << true }
end

worker_count.times do |i|
  worker_pid = fork do
    Process.setproctitle("fooworker: worker")
    puts("Started worker process: #{Process.pid}")

    loop do
      break if !shutdown_queue.empty?

      # Perform the desired tasks inside the loop
      # ...
    end
  end

  worker_pids << worker_pid
end

worker_pids.each { |pid| Process.waitpid(pid) }
