require 'optparse'
require 'net/http'
require 'json'
require 'openssl'

RequiredOpts = [:task, :host, :port, :username, :password]
SupportedTasks = ['list-channels', 'create-channel', 'update-channel', 'delete-channel', 'show-channel']

def parse_opts(args)
  options = {}
  opt = OptionParser.new do |opts|
    opts.banner = "Usage: eywa_tools.rb [options]"

    opts.on("-?", "--help", "Prints this help") do
      puts opts
      exit
    end

    opts.on("-t task", "--task TASK",
            "Specify the task") do |task|
      options[:task] = task
    end

    opts.on("-h host", "--host HOST",
            "Host name of Eywa") do |host|
      options[:host] = host
    end

    opts.on("-p port", "--port PORT",
            "Port number of Eywa") do |port|
      options[:port] = port
    end

    opts.on("-u username", "--username USERNAME",
            "Admin username") do |username|
      options[:username] = username
    end

    opts.on("-w password", "--password PASSWORD",
            "Admin password") do |password|
      options[:password] = password
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
    puts "Supported tasks are: #{SupportedTasks.join(',')}."
    exit 1
  end

  options
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
      auth = JSON.parse(response.body)["auth_token"]
    end
  rescue => e
    puts e.message
  end

  if code.to_i != 200 || auth.length == 0
    puts "Login failed. returned error code: #{code}. please check your username and password."
    exit 1
  end
  puts 'Login successful!'
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
  channel_id = nil
  if opt[:channel_id].nil? || opt[:channel_id].length == 0
    list_channels(opt)
    print "Input channel id to continue: "
    channel_id = gets.chomp
    if channel_id.length == 0
      puts 'Empty channel id, quitting...'
      exit 1
    end
  else
    channel_id = opt[:channel_id]
  end

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
  yes_or_no = gets.chomp
  if yes_or_no != 'yes'
    puts 'Nothing is deleted.'
    exit 0
  end

  list_channels(opt)
  print "Input channel id to continue: "
  channel_id = gets.chomp
  if channel_id.length == 0
    puts 'Empty channel id, quitting...'
    exit 1
  end

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
  params[:name] = gets.chomp

  print "Description (required): "
  params[:description] = gets.chomp

  print "Tags (optional, separate tags with [,]): "
  params[:tags] = gets.chomp.split(',')

  print "Fields (required, example: temperature:float,brightness:int): "
  params[:fields] = gets.chomp.split(',').inject({}) do |map, field|
    map[field.split(":").first] = field.split(":").last
    map
  end

  print "AccessToken (required, separate access tokens with [,]): "
  params[:access_tokens] = gets.chomp.split(',')

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
  yes_or_no = gets.chomp
  if yes_or_no != 'yes'
    puts 'Nothing is updated.'
    exit 0
  end

  list_channels(opt)
  print "Input channel id to continue: "
  channel_id = gets.chomp
  if channel_id.length == 0
    puts 'Empty channel id, quitting...'
    exit 1
  end

  puts 'Current channel definition:'
  opt[:channel_id] = channel_id
  show_channel(opt)

  params = {}
  print "Name (press enter to skip): "
  params[:name] = gets.chomp
  params.delete(:name) if params[:name].nil? || params[:name].length == 0

  print "Description (press enter to skip): "
  params[:description] = gets.chomp
  params.delete(:description) if params[:description].nil? || params[:description].length == 0

  print "Tags (optional, separate tags with [,]. press enter to skip): "
  params[:tags] = gets.chomp.split(',')
  params.delete(:tags) if params[:tags].nil? || params[:tags].length == 0

  print "Fields (required, example: temperature:float,brightness:int. press enter to skip): "
  params[:fields] = gets.chomp.split(',').inject({}) do |map, field|
    map[field.split(":").first] = field.split(":").last
    map
  end
  params.delete(:fields) if params[:fields].nil? || params[:fields].length == 0

  print "AccessToken (required, separate access tokens with [,], press enter to skip): "
  params[:access_tokens] = gets.chomp.split(',')
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
else
  puts "Unsupported task: [#{options[:task]}]."
  exit 1
end
