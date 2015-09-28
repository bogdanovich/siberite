# gem memcache-client
require 'memcache'

total_elements = ARGV[0].to_i

client = MemCache.new 'localhost:22133'
client.delete 'test'
start_time = Time.now
total_elements.times do |i|
  client.set('test', i, 0, true)
end
puts "#{total_elements} inserted in #{Time.now - start_time} seconds"

# Insert loop oneliner
# require 'memcache'; client = MemCache.new 'localhost:22133'; client.stats ; loop { client.set('test', rand(100_000_000), 0, true) }
