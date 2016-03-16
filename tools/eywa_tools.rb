require 'optparse'
require 'net/http'
require 'json'
require 'openssl'

RequiredOpts = [:task, :host, :port, :username, :password]
SupportedTasks = [
  'list-channels', 'create-channel', 'update-channel',
  'delete-channel', 'show-channel', 'connection-count',
  'connection-status', 'show-settings', 'update-settings',
  'send-to-connection', 'request-to-connection', 'query-value',
  'query-series', 'query-raw'
]

def parse_opts(args)
  options = {}
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

    opts.on("-s", "--use-ssl",
            "Use SSL") do
      options[:use_ssl] = true
    end

  end

  opt.parse!(args) rescue puts opt

  RequiredOpts.each do |arg|
    if options[arg].nil? || options[arg].length == 0
      puts "Not enough options. required options are: #{RequiredOpts}."
      exit 1
    end
  end

  unless SupportedTasks.include?(options[:task])
    puts "Unsupported task: #{options[:task]}."
    puts "Supported tasks are: \n  #{SupportedTasks.each_slice(3).map{|slice| slice.join(', ')}.join("\n  ")}."
    exit 1
  end

  options
end

def get_channel_id(opt)
  channel_id = nil
  if opt[:channel_id].nil? || opt[:channel_id].length == 0
    list_channels(opt)
    print "Input channel id to continue: "
    channel_id = gets.chomp.strip
    if channel_id.length == 0
      puts 'Empty channel id, quitting...'
      exit 1
    end
    opt[:channel_id] = channel_id
  else
    channel_id = opt[:channel_id]
  end

  channel_id
end

def get_device_id(opt)
  device_id = nil
  if opt[:device_id].nil? || opt[:device_id].length == 0
    print "Input device id to continue: "
    device_id = gets.chomp.strip
    if device_id.length == 0
      puts 'Empty device id, quitting...'
      exit 1
    end
    opt[:device_id] = device_id
  else
    device_id = opt[:device_id]
  end

  device_id
end

def get_message(opt)
  message = nil
  if opt[:message].nil? || opt[:message].length == 0
    print "Input message to continue: "
    message = gets.chomp.strip
    if message.length == 0
      puts 'Empty message, quitting...'
      exit 1
    end
    opt[:message] = message
  else
    message = opt[:message]
  end

  message
end

######################### Tasks #########################
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
      auth = JSON.parse(response.body)["auth_token"]
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200 || auth.length == 0
    puts "Login failed. returned error code: #{code}. please check your username and password."
    exit 1
  end
  # puts 'Login successful!'
  return auth
end

def list_channels(opt)
  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      puts JSON.pretty_generate(JSON.parse(response.body))
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts 'Failed to list channels.'
    exit 1
  end
end

def show_channel(opt)
  channel_id = get_channel_id(opt)

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}")
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      puts JSON.pretty_generate(JSON.parse(response.body))
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts 'Failed to show channel details.'
    exit 1
  end
end

def delete_channel(opt)
  print "Are you sure you want to delete a channel?(yes/no): "
  yes_or_no = gets.chomp.strip
  if yes_or_no != 'yes'
    puts 'Nothing is deleted.'
    exit 0
  end

  channel_id = get_channel_id(opt)

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}")
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
    puts "Failed to delete channel: #{channel_id}."
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
    puts 'Failed to create channel.'
    exit 1
  end

  resp = JSON.parse(resp)
  puts "Channel created with id: #{resp['id']}"
  opt[:channel_id] = resp['id']
  show_channel(opt)
end

def update_channel(opt)
  print "Are you sure you want to update a channel?(yes/no): "
  yes_or_no = gets.chomp.strip
  if yes_or_no != 'yes'
    puts 'Nothing is updated.'
    exit 0
  end

  channel_id = get_channel_id(opt)

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
    puts 'Failed to update channel.'
    exit 1
  end

  puts "Channel updated with id: #{channel_id}"
  show_channel(opt)
end

def connection_count(opt)
  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/connections/count")
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
    puts 'Failed to get connection count.'
    exit 1
  end
end

def connection_status(opt)
  channel_id = get_channel_id(opt)
  device_id = get_device_id(opt)

  print "With connection history?(yes/no): "
  with_history = gets.chomp.strip
  if with_history != 'yes'
    puts 'Skipping connection history...'
    with_history = false
  else
    with_history = true
  end

  code = nil
  uri = URI("http#{opt[:use_ssl] ? 's': ""}://#{opt[:host]}:#{opt[:port]}/admin/channels/#{channel_id}/devices/#{device_id}/status")
  uri.query = URI.encode_www_form({history: with_history})
  begin
    Net::HTTP.start(uri.host, uri.port, :use_ssl => uri.scheme == 'https', :verify_mode => OpenSSL::SSL::VERIFY_NONE) do |http|
      request = Net::HTTP::Get.new uri
      request.add_field('Authentication', opt[:auth_token])

      response = http.request request
      code = response.code
      puts JSON.pretty_generate(JSON.parse(response.body))
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts 'Failed to get connection status.'
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
      puts JSON.pretty_generate(JSON.parse(response.body))
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts 'Failed to get Eywa settings.'
    exit 1
  end
end

def send_to_connection(opt)
  channel_id = get_channel_id(opt)
  device_id = get_device_id(opt)
  message = get_message(opt)

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
    puts "Failed to send message: [#{message}] to device: [#{device_id}] in channel: [#{channel_id}]."
    exit 1
  end
  puts 'Message sent successfully!'
end

def request_to_connection(opt)
  channel_id = get_channel_id(opt)
  device_id = get_device_id(opt)
  message = get_message(opt)

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
    puts "Failed to request message: [#{message}] to device: [#{device_id}] in channel: [#{channel_id}]."
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
      puts JSON.pretty_generate(JSON.parse(response.body))
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200
    puts 'Failed to update Eywa settings.'
    exit 1
  end
  puts 'Successfully updated Eywa settings!'
end

def query_value(opt)
end

def query_series(opt)
end

def query_raw(opt)
end

Signal.trap("INT") {
  puts "\nTask aborted."
  exit 1
}

options = parse_opts(ARGV)
options[:auth_token] = login(options)

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
when 'connection-count'
  connection_count(options)
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
when 'send-to-connection'
  send_to_connection(options)
when 'request-to-connection'
  request_to_connection(options)
else
  puts "Unsupported task: [#{options[:task]}]."
  exit 1
end
