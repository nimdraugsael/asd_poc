namespace :save_results_worker do
  desc "Start the forking worker"
  task :start do
    exec "ruby lib/workers/db_worker.rb"
  end

  desc "Stop the forking worker"
  task :stop do
    master_pid_file = './tmp/db_workers/master.pid'
    if File.exist?(master_pid_file)
      master_pid = File.read(master_pid_file).to_i
      Process.kill('TERM', master_pid)
    else
      puts "No master worker process found"
    end
  end
end
