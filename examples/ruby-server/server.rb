require 'webrick'

port = ENV['PORT']&.to_i || 3000

server = WEBrick::HTTPServer.new(
  Port: port,
  DocumentRoot: '.'
)

server.mount_proc '/' do |req, res|
  res.body = <<~HTML
    <h1>Ruby WEBrick Server</h1>
    <p>Port: #{port}</p>
    <p>Host: #{req.header['host'].first}</p>
    <p>Path: #{req.path}</p>
  HTML
  res['Content-Type'] = 'text/html'
end

puts "Ruby WEBrick server listening on port #{port}"
server.start
