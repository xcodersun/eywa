require 'optparse'
require 'net/http'
require 'json'
require 'openssl'

LOAD_MSG =<<-D
Gem `websocket-client-simple` is not installed.
Tasks: [attach-connection, tail-log] will be disabled.
To install:
  `gem install websocket-client-simple`

D

RequiredOpts = [:task, :host, :port, :username, :password]
SupportedTasks = [
  'list-channels', 'create-channel', 'update-channel',
  'delete-channel', 'show-channel', 'connection-counts',
  'connection-status', 'show-settings', 'update-settings',
  'send-to-connection', 'request-to-connection', 'query-value',
  'query-series', 'query-raw', 'scan-connections', 'tail-log',
  'attach-connection'
]

def load_libs
  begin
    require('websocket-client-simple')
  rescue LoadError
    puts LOAD_MSG
    exit 1
  end
end

def parse_opts(args)
  options = {nop: true}
  opt = OptionParser.new do |opts|
    opts.banner = <<-USG
Supported Tasks:
  #{SupportedTasks.each_slice(3).map{|slice| slice.join(', ')}.join("\n  ")}
Usage: eywa_tools.rb [options]
    USG

    opts.on("-?", "--help", "Prints this help") do
      puts opts
      exit
    end

    opts.on("-t task", "--task",
            "Specify the task") do |task|
      options[:task] = task
    end

    opts.on("-h host", "--host",
            "Host name of Eywa") do |host|
      options[:host] = host
    end

    opts.on("-p port", "--port",
            "Port number of Eywa") do |port|
      options[:port] = port
    end

    opts.on("-u username", "--username",
            "Admin username") do |username|
      options[:username] = username
    end

    opts.on("-w password", "--password",
            "Admin password") do |password|
      options[:password] = password
    end

    opts.on("-c channel_id", "--channel-id",
            "Channel ID") do |channel_id|
      options[:channel_id] = channel_id
    end

    opts.on("-d device_id", "--device-id",
            "Device ID") do |device_id|
      options[:device_id] = device_id
    end

    opts.on("-m message", "--message",
            "Message sent to device") do |message|
      options[:message] = message
    end

    opts.on("-o timeout", "--timeout",
            "Timeout for request message") do |timeout|
      options[:timeout] = timeout
    end

    opts.on("-f field", "--field",
            "Field in query") do |field|
      options[:field] = field
    end

    opts.on("-g tags", "--tags",
            "Tags in query") do |tags|
      options[:tags] = tags
    end

    opts.on("-T time_range", "--time-range",
            "Time Range in query") do |time_range|
      options[:time_range] = time_range
    end

    opts.on("-i time_interval", "--time-interval",
            "Time Interval in query") do |time_interval|
      options[:time_interval] = time_interval
    end

    opts.on("-a aggregation", "--aggregation",
            "Aggregation in query") do |aggregation|
      options[:aggregation] = aggregation
    end

    opts.on("-s", "--use-ssl",
            "Use SSL") do
      options[:use_ssl] = true
    end

    opts.on("-y", "--yes",
            "Say yes") do
      options[:yes] = 'yes'
    end

    opts.on("-N", "--nop-false",
            "Turn off nop in raw query, defaults turned on") do
      options[:nop] = false
    end

  end

  opt.parse!(args) rescue puts opt

  RequiredOpts.each do |arg|
    if options[arg].nil? || options[arg].length == 0
      puts "Not enough options. required options are: #{RequiredOpts}."
      puts
      puts opt
      exit 1
    end
  end

  unless SupportedTasks.include?(options[:task])
    puts "Unsupported task: #{options[:task]}."

    puts
    puts opt
    exit 1
  end

  options
end

def get_option(opt, option_name, skip=false, pre_hook=nil, post_hook=nil)
  if opt[option_name].nil? || opt[option_name].length == 0
    pre_hook.call(opt, option_name) if pre_hook
    print "Input #{option_name.to_s.gsub('_', ' ')} to continue: "
    option_value = gets.chomp.strip
    if option_value.length == 0 && !skip
      puts "Empty #{option_name.to_s.gsub('_', ' ')}, quitting..."
      exit 1
    end
    opt[option_name] = option_value
  end

  opt[option_name] = post_hook.call(opt, option_name, opt[option_name]) if post_hook

  opt[option_name]
end

def print_wiki_query
  puts 'For detailed query syntax, please refer to https://github.com/xcodersun/eywa/wiki/Query-Eywa .'
end

def print_response(response)
  puts 'Response:'
  puts JSON.pretty_generate(JSON.parse(response))
rescue => e
  puts response
end

def login(opt)
  auth = nil
  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/login")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.basic_auth opt[:username], opt[:password]


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

######################### Tasks #########################

def list_channels(opt)
  code = nil
  resp = ""
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels")
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
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}")
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

def delete_channel(opt)
  if opt[:yes].nil?
    print "Are you sure you want to delete a channel?(yes/no): "
    yes_or_no = gets.chomp.strip
    if yes_or_no != 'yes'
      puts 'Nothing is deleted.'
      exit 0
    end
  end

  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})
  print "Do you want to delete all the indices belong to this channel?(yes/no): "
  with_indices = gets.chomp.strip == 'yes' ? true : false

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}")
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

def create_channel(opt)
  params = {}
  print "Name (required): "
  params[:name] = gets.chomp.strip

  print "Description (required): "
  params[:description] = gets.chomp.strip

  print "Tags (optional, separate tags with [,]): "
  params[:tags] = gets.chomp.strip.split(',')

  print "Fields (required, example: temperature:float,brightness:int): "
  params[:fields] = gets.chomp.strip.split(',').inject({}) do |map, field|
    map[field.split(":").first] = field.split(":").last
    map
  end

  print "AccessToken (required, separate access tokens with [,]): "
  params[:access_tokens] = gets.chomp.strip.split(',')

  puts "Please review your channel:"
  puts JSON.pretty_generate(params)

  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Post.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = params.to_json

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
  if opt[:yes].nil?
    print "Are you sure you want to update a channel?(yes/no): "
    yes_or_no = gets.chomp.strip
    if yes_or_no != 'yes'
      puts 'Nothing is updated.'
      exit 0
    end
  end

  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})

  puts 'Current channel definition:'
  show_channel(opt)

  params = {}
  print "Name (press enter to skip): "
  params[:name] = gets.chomp.strip
  params.delete(:name) if params[:name].nil? || params[:name].length == 0

  print "Description (press enter to skip): "
  params[:description] = gets.chomp.strip
  params.delete(:description) if params[:description].nil? || params[:description].length == 0

  print "Tags (optional, separate tags with [,]. press enter to skip): "
  params[:tags] = gets.chomp.strip.split(',')
  params.delete(:tags) if params[:tags].nil? || params[:tags].length == 0

  print "Fields (required, example: temperature:float,brightness:int. press enter to skip): "
  params[:fields] = gets.chomp.strip.split(',').inject({}) do |map, field|
    map[field.split(":").first] = field.split(":").last
    map
  end
  params.delete(:fields) if params[:fields].nil? || params[:fields].length == 0

  print "AccessToken (required, separate access tokens with [,], press enter to skip): "
  params[:access_tokens] = gets.chomp.strip.split(',')
  params.delete(:access_tokens) if params[:access_tokens].nil? || params[:access_tokens].length == 0

  puts "Please review changes to your channel:"
  puts JSON.pretty_generate(params)

  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Put.new uri
      request.add_field('Authentication', opt[:auth_token])
      request.body = params.to_json

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

def connection_counts(opt)
  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/connections/counts")
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
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})
  device_id = get_option(opt, :device_id, false)

  if opt[:yes].nil?
    print "With connection history?(yes/no): "
    yes_or_no = gets.chomp.strip
    if yes_or_no != 'yes'
      puts 'Skipping connection history...'
      yes_or_no = false
    else
      yes_or_no = true
    end
  else
    yes_or_no = true
  end

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/devices/#{device_id}/status")
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

def send_to_connection(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})
  device_id = get_option(opt, :device_id)
  message = get_option(opt, :message)

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/devices/#{device_id}/send")
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
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})
  device_id = get_option(opt, :device_id)
  message = get_option(opt, :message)

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/devices/#{device_id}/request")
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

def scan_connections(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt| list_channels(opt)})
  params = {
    size: get_option(opt, :size, true),
    last: get_option(opt, :last, true)
  }.delete_if{|_, v| v.length == 0}

  puts "Please review your query:"
  puts JSON.pretty_generate(params)
  puts "Press enter to continue, or Ctrl-C to abort"
  gets

  code = nil
  resp = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/connections/scan")
  uri.query = URI.encode_www_form(params)
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

def attach_connection(opt)
  channel_id = get_option(opt, :channel_id, false, Proc.new{|opt, _| list_channels(opt)})
  device_id = get_option(opt, :device_id, false)

  url = "ws#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/devices/#{device_id}/attach"
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

def _sleep
  begin
    sleep
  rescue Interrupt
  end
end

def cleanup_ws(opt)
  opt[:ws].close if opt[:ws]
end

options = parse_opts(ARGV)
options[:auth_token] = login(options)

Signal.trap("INT") {
  puts "\nTask aborted."
  exit 1
}

case options[:task]
when 'list-channels'
  list_channels(options)
when 'show-channel'
  show_channel(options)
when 'delete-channel'
  delete_channel(options)
when 'create-channel'
  create_channel(options)
when 'update-channel'
  update_channel(options)
when 'connection-counts'
  connection_counts(options)
when 'connection-status'
  connection_status(options)
when 'show-settings'
  show_settings(options)
when 'update-settings'
  update_settings(options)
when 'query-value'
  query_value(options)
when 'query-series'
  query_series(options)
when 'query-raw'
  query_raw(options)
when 'scan-connections'
  scan_connections(options)
when 'send-to-connection'
  send_to_connection(options)
when 'request-to-connection'
  request_to_connection(options)
when 'tail-log'
  load_libs
  tail_log(options)
when 'attach-connection'
  load_libs
  attach_connection(options)
else
  puts "Unsupported task: [#{options[:task]}]."
  exit 1
end
