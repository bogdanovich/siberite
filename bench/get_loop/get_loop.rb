# gem memcache-client
require 'memcache'

total_elements = ARGV[0].to_i

client = MemCache.new 'localhost:22133'

start_time = Time.now

total_elements.times do
  client.get('test', true)
end

puts "#{total_elements} keys read in #{Time.now - start_time} seconds"

# Get loop oneliner
# require 'memcache'; client = MemCache.new 'localhost:22133'; client.stats ; loop { client.get('test', true) }
