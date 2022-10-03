require 'thor'

module Backup
  # The backup Command Line Interface (CLI entry point, see `bin/backup`)
  class CLI < Thor
    desc 'list_devices', 'list the devices available on the current system'
    def list_devices
      puts 'TODO'
    end
  end
end
