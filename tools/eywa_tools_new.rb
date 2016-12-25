require 'optparse'
require 'net/http'
require 'json'
require 'openssl'

$profile = {}

def print_response(response)
  puts 'Response:'
  puts JSON.pretty_generate(JSON.parse(response))
rescue => e
  puts response
end

def get_option(opt, option_name)
  if opt[option_name].nil? || opt[option_name].length == 0
    puts "Missing #{option_name}, quitting..."
    exit 1
  end
  opt[option_name]
end

def load_libs
  begin
    require('websocket-client-simple')
  rescue LoadError
    puts LOAD_MSG
    exit 1
  end
end

def _sleep
  begin
    sleep
  rescue Interrupt
  end
end

def login
  auth = nil
  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/login")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.basic_auth $profile['username'], $profile['password']
      response = http.request request
      code = response.code
      auth = JSON.parse(response.body)["auth_token"] rescue ""
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200 || auth.length == 0
    puts "Login failed. returned error code: #{code}. please check your username and password."
    exit 1
  end

  return auth
end

def list_channels(opt)
  code = nil
  resp = ""
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to list channels. code=#{code}"
    exit 1
  end
end

def show_channel(opt)
  channel_id = get_option(opt, :channel_id)

  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to show channel details. code=#{code}"
    exit 1
  end
end

def create_channel(opt)
  file = File.read(opt[:template])
  begin
    channel = JSON.parse(file)
  rescue JSON::ParserError => e
    puts "Fail to parse #{opt[:template]}, please check the json format"
    exit 1
  end

  code = nil
  resp = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Post.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = channel.to_json

      response = http.request request
      code = response.code
      resp = response.body
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 201
    puts resp
    puts "Failed to create channel. code=#{code}"
    exit 1
  end

  resp = JSON.parse(resp)
  puts "Channel created with id: #{resp['id']}"
  opt[:channel_id] = resp['id']
  show_channel(opt)
end

def update_channel(opt)
  channel_id = get_option(opt, :channel_id)

  puts 'Current channel definition:'
  show_channel(opt)
  print "Are you sure you want to update channel it? (yes/no): "
  yes_or_no = gets.chomp.strip
  if yes_or_no != 'yes'
    puts 'Nothing is updated.'
    exit 0
  end

  file = File.read(opt[:template])
  begin
    channel = JSON.parse(file)
  rescue JSON::ParserError => e
    puts "Fail to parse #{opt[:template]}, please check the json format"
    exit 1
  end

  puts "Please review changes to your channel:"
  puts JSON.pretty_generate(channel)

  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Put.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = channel.to_json

      response = http.request request
      code = response.code
      puts response.body
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to update channel. code=#{code}"
    exit 1
  end

  puts "Channel updated with id: #{channel_id}"
  show_channel(opt)
end

def delete_channel(opt)
  puts 'Current channel definition:'
  show_channel(opt)
  print "Are you sure you want to delete it? (yes/no): "
  yes_or_no = gets.chomp.strip
  if yes_or_no != 'yes'
    puts 'Nothing is deleted.'
    exit 0
  end

  channel_id = opt[:channel_id]
  print "Do you want to delete all the indices belong to this channel?(yes/no): "
  with_indices = gets.chomp.strip == 'yes' ? true : false

  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}")
  uri.query = URI.encode_www_form({with_indices: with_indices})
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Delete.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to delete channel: #{channel_id}. code=#{code}"
    exit 1
  end

  puts "Channel: #{channel_id} is successfully deleted."
  list_channels(opt)
end

def connection_counts(opt)
  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/connections/counts")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      puts response.body
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to get connection counts. code=#{code}"
    exit 1
  end
end

def connection_status(opt)
  channel_id = get_option(opt, :channel_id)
  device_id = get_option(opt, :device_id)

  print "With connection history?(yes/no): "
  yes_or_no = gets.chomp.strip
  if yes_or_no != 'yes'
    puts 'Skipping connection history...'
    yes_or_no = false
  else
    yes_or_no = true
  end

  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}/devices/#{device_id}/status")
  uri.query = URI.encode_www_form({history: yes_or_no})
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to get connection status. code=#{code}"
    exit 1
  end
end

def send_to_connection(opt)
  channel_id = get_option(opt, :channel_id)
  device_id = get_option(opt, :device_id)
  message = get_option(opt, :message)

  code = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}/devices/#{device_id}/send")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Post.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = message

      response = http.request request
      code = response.code
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to send message: [#{message}] to device: [#{device_id}] in channel: [#{channel_id}]. code=#{code}"
    exit 1
  end
  puts 'Message sent successfully!'
end

def request_to_connection(opt)
  channel_id = get_option(opt, :channel_id)
  device_id = get_option(opt, :device_id)
  message = get_option(opt, :message)

  code = nil
  resp = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}/devices/#{device_id}/request")
  uri.query = URI.encode_www_form({timeout: opt[:timeout]}) if !opt[:timeout].nil? && opt[:timeout].length > 0
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Post.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = message

      response = http.request request
      code = response.code
      resp = response.body
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to request message: [#{message}] to device: [#{device_id}] in channel: [#{channel_id}]. code=#{code}"
    exit 1
  end

  puts 'Message request successfully!'
  puts "Response: \n  #{resp}"
end

def scan_connections(opt)
  channel_id = get_option(opt, :channel_id)

  code = nil
  resp = nil
  uri = URI("#{$profile['protocol']}://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}/connections/scan")

  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      http.read_timeout = 3600
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      resp = response.body
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts resp
    puts "Failed to query connections. code=#{code}"
    print_wiki_query
    exit 1
  end

  print_response(resp)
  puts 'Successfully scanned connections!'
end

def attach_connection(opt)
  channel_id = get_option(opt, :channel_id)
  device_id = get_option(opt, :device_id)

  url = "ws://#{$profile['host']}:#{$profile['port']}/admin/channels/#{channel_id}/devices/#{device_id}/attach"
  ws = WebSocket::Client::Simple.connect(url, {headers: {"Authentication"=>opt[:auth_token]}})
  opt[:ws] = ws
  welcome = nil
  ws.on(:message) do |msg|
    welcome = msg unless welcome
    puts msg.data
  end
  sleep 1
  if welcome.nil?
    puts "Unable to attach to connection, check if it's online by `connection-status` ..."
  else
    _sleep
  end
end

##########################
#          Main          #
##########################
USAGE_BANNER = "Usage: eywa_tools.rb [subcommand [options]]"

SUBUSAGE = <<HELP
Subcommands are:
   channel    :     create, update, list, delete and show channels, etc.
   connection :     count, show, send to and request for connection, etc.
See 'eywa_tools.rb SUBCOMMAND --help' for more information on a specific command.
HELP

PROFILE_FILE = "./profile/profile.json"

def usage_menu
  puts USAGE_BANNER
  puts
  puts SUBUSAGE
end

def read_profile
  if File.exist?(PROFILE_FILE)
    file = File.read(PROFILE_FILE)
    $profile = JSON.parse(file)
  end
end

read_profile
if $profile.empty?
  puts "Please create a profile."
  exit 1
end

options = {nop: true}

subcommands = { 
  'channel' => OptionParser.new do |opts|
    opts.banner = "Usage eywa_tools channels [options]"
    opts.separator ""
    opts.on("-c template.json", "--create", "Create a new channel") do |template|
      options[:subcommand] = "create_channel"
      options[:template] = template
    end

    opts.on("-u template.json", "--update", "Update a channel") do |template|
      options[:subcommand] = "update_channel"
      options[:template] = template
    end

    opts.on("-l", "--list", "List all channels") do
      options[:subcommand] = "list_channels"
    end

    opts.on("-s channel_id", "--show", "Show a channel") do |channel_id|
      options[:subcommand] = "show_channel"
      options[:channel_id] = channel_id
    end

    opts.on("-d channel_id", "--delete", "Delete a channel") do |channel_id|
      options[:subcommand] = "delete_channel"
      options[:channel_id] = channel_id
    end

    opts.on("-i channel_id", "--id", "Specify a channel by ID") do |channel_id|
      options[:channel_id] = channel_id
    end

   end,

  'connection' => OptionParser.new do |opts|
    opts.banner = "Usage eywa_tools connection [options]"
    opts.separator ""

    opts.on("-c", "--counts", "Count total connections") do
      options[:subcommand] = "connection_counts"
    end

    opts.on("-u", "--status", "Check  connection status") do
      options[:subcommand] = "connection_status"
    end

    opts.on("-s message", "--send", "Send a message to device") do |message|
      options[:subcommand] = "send_to_connection"
      options[:message] = message
    end

    opts.on("-r message", "--request", "Request a message from device") do |message|
      options[:subcommand] = "request_to_connection"
      options[:message] = message
    end

    opts.on("-a", "--attach", "Attach to a connection") do
      options[:subcommand] = "attach_connection"
    end

    opts.on("-n", "--scan", "scan channel connections") do
      options[:subcommand] = "scan_connections"
    end

    opts.on("-i channel_id", "--channel", "Specify a channel by ID") do |channel_id|
      options[:channel_id] = channel_id
    end

    opts.on("-d device_id", "--device", "Specify a device by ID") do |device_id|
      options[:device_id] = device_id
    end

   end
 }

command = ARGV.shift

if subcommands[command].nil?
  puts
  usage_menu
  exit 1
else
  begin subcommands[command].parse!
  rescue OptionParser::MissingArgument
    puts "Missing option for command \"#{command}\""
    puts "Please run: eway_tools #{command} --help"
    puts
    exit 1
  rescue OptionParser::InvalidOption
    puts "Invalid option for command \"#{command}\""
    puts "Please run: eway_tools #{command} --help"
    puts
    exit 1
  end
  subcommands[command].order!
end

Signal.trap("INT") {
  puts "\nTask aborted."
  exit 1
}

options[:auth_token] = login
case options[:subcommand]
when 'create_channel'
  create_channel(options)
when 'update_channel'
  update_channel(options)
when 'list_channels'
  list_channels(options)
when 'show_channel'
  show_channel(options)
when 'delete_channel'
  delete_channel(options)
when 'connection_counts'
  connection_counts(options)
when 'connection_status'
  connection_status(options)
when 'send_to_connection'
  send_to_connection(options)
when 'request_to_connection'
  request_to_connection(options)
when 'attach_connection'
  load_libs
  attach_connection(options)
when 'scan_connections'
  scan_connections(options)
end

=begin

def show_settings(opt)
  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/configs")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to get Eywa settings. code=#{code}"
    exit 1
  end
end

def update_settings(opt)
  puts 'Please input settings in a flattened fashion, example: connections.websocket.timeouts.read:4s .'
  puts 'Multiple settings are delimited by comma.'
  puts '-----------------------------------------'
  settings = gets.chomp.strip
  settings = settings.split(',').map(&:strip)
  updates = settings.inject({}) do |hash, setting|
    key, value = setting.split(':')
    nestings = key.split('.')
    root = hash
    nestings.each_with_index do |nest, idx|
      if hash[nest].nil?
        if idx == nestings.count - 1
          hash[nest] = value
        else
          hash[nest] = {}
          hash = hash[nest]
        end
      else
        hash = hash[nest]
      end
    end
    root
  end

  puts '-----------------------------------------'
  puts 'Please review your changes.'
  puts JSON.pretty_generate(updates)

  puts
  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/configs")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Put.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = updates.to_json

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts "Failed to update Eywa settings. code=#{code}"
    exit 1
  end
  puts 'Successfully updated Eywa settings!'
end

def query_value(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt| list_channels(opt)})
  show_channel(opt)
  params = {
    field: get_option(opt, :field),
    tags: get_option(opt, :tags, true, nil, Proc.new{|_, _, tags|
      tags.split(',').map(&:strip).join(',')
    }),
    summary_type: get_option(opt, :aggregation),
    time_range: get_option(opt, :time_range)
  }.delete_if{|_, v| v.length == 0}

  puts "Please review your query:"
  puts JSON.pretty_generate(params)
  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/value")
  uri.query = URI.encode_www_form(params)
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      http.read_timeout = 3600
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts resp
    puts "Failed to query value. code=#{code}"
    print_wiki_query
    exit 1
  end

  puts 'Successfully queried value!'
end

def query_series(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt| list_channels(opt)})
  show_channel(opt)
  params = {
    field: get_option(opt, :field),
    tags: get_option(opt, :tags, true, nil, Proc.new{|_, _, tags|
      tags.split(',').map(&:strip).join(',')
    }),
    summary_type: get_option(opt, :aggregation),
    time_range: get_option(opt, :time_range),
    time_interval: get_option(opt, :time_interval, false)
  }.delete_if{|_, v| v.length == 0}

  puts "Please review your query:"
  puts JSON.pretty_generate(params)
  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/series")
  uri.query = URI.encode_www_form(params)
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      http.read_timeout = 3600
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts resp
    puts "Failed to query series. code=#{code}"
    print_wiki_query
    exit 1
  end

  puts 'Successfully queried series!'
end

def query_raw(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt| list_channels(opt)})
  show_channel(opt)
  params = {
    tags: get_option(opt, :tags, true, nil, Proc.new{|_, _, tags|
      tags.split(',').map(&:strip).join(',')
    }),
    time_range: get_option(opt, :time_range),
  }.delete_if{|_, v| v.length == 0}.merge(nop: opt[:nop])

  puts "Please review your query:"
  puts JSON.pretty_generate(params)
  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/raw")
  uri.query = URI.encode_www_form(params)
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      http.read_timeout = 3600
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      print_response(response.body)
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts resp
    puts "Failed to query series. code=#{code}"
    print_wiki_query
    exit 1
  end

  if opt[:nop]
    puts "Successfully queried raw in nop=true mode! To turn if off, please use '-N' option."
  else
    puts 'Successfully queried raw!'
  end
end

def tail_log(opt)
  url = "ws#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/tail"
  ws = WebSocket::Client::Simple.connect(url, {headers: {"Authentication"=>opt[:auth_token]}})
  opt[:ws] = ws
  welcome = nil
  ws.on(:message) do |msg|
    welcome = msg unless welcome
    puts msg.data
  end
  sleep 1
  if welcome.nil?
    puts "Unable to tail server log, something weird happened ..."
  else
    _sleep
  end
end

def cleanup_ws(opt)
  opt[:ws].close if opt[:ws]
end

options = parse_opts(ARGV)
  
=end
