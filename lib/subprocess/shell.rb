require 'open3'

# rubocop:disable Style/HashSyntax
module Subprocess
  # improve security by only allowing safest variants
  module Shell
    # returns merged stdout_stderr, status
    def self.exec(command, untrusted = [], dir: nil, env: nil)
      untrusted = [untrusted] unless untrusted.is_a? Array
      out_err, status = do_exec(command, untrusted, dir: dir, env: env)
      p out_err if %w[1 true on].include?(ENV['DEBUG'])
      [out_err.strip, status]
    end

    # rubocop:disable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity, Metrics/MethodLength
    private_class_method def self.do_exec(command, untrusted, dir: nil, env: nil)
      with_clean_env do
        if dir && env
          Open3.capture2e(env, command, *untrusted.map(&:to_s), chdir: dir)
        elsif dir
          Open3.capture2e(command, *untrusted.map(&:to_s), chdir: dir)
        elsif env
          Open3.capture2e(env, command, *untrusted.map(&:to_s))
        else
          Open3.capture2e(command, *untrusted.map(&:to_s))
        end
      end
    end
    # rubocop:enable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity, Metrics/MethodLength

    private_class_method def self.with_clean_env # rubocop:disable Metrics/MethodLength
      backup = ENV.to_h
      new_hsh = ENV.to_h
      # only keep certain ENV variables
      to_keep = %w[SHELL PWD HOME LANG USER TERM _ PATH rvm_path]
      clean_env = new_hsh.select do |k, _|
        to_keep.include?(k) ||
          k.start_with?('RUBY') ||
          # keep all BUNDLE variables except for Gemfile (it's the one from doctor)
          (k.start_with?('BUNDLE') && k != 'BUNDLE_GEMFILE') ||
          k.start_with?('GEM')
      end
      ENV.replace(clean_env)
      yield
    ensure
      ENV.replace(backup)
    end
  end
end
# rubocop:enable Style/HashSyntax

# Patching Kernel's ` backtick command because we want to prohibit usage of it.
module Kernel
  def `(_cmd)
    raise SecurityError, 'Usage of backticks is prohibited. Use the provided Shell class to make system calls instead.'
  end
end

class Thor
  module Shell
    # patching Thor methods that originally use backticks
    class Basic
      # original implementation uses backticks which we're prohibiting
      def dynamic_width_stty
        IO.popen('stty size').read.split[1].to_i
      end

      # original implementation uses backticks which we're prohibiting
      def dynamic_width_tput
        IO.popen('tput cols').read.to_i
      end
    end
  end
end
