require 'csv'
require 'open-uri'
require 'json'

ids = [151370911, 97542154]

json = ids.map do |id|
  puts "requesting #{id}"
  resp = open("http://api.soundcloud.com/tracks/#{id}.json?client_id=1182e08b0415d770cfb0219e80c839e8").read
  r = JSON.parse(resp)
  puts "done requesting #{id}"
  r
end

data = CSV.read("./us-ag-productivity-2004.csv")

header = data.shift

header << "Playback count" << "Title" << "Link" << "Artwork"

new_data = data.map do |d|
  puts "appending to #{d[0]}"
  song = json.sample
  d << song["playback_count"]
  d << song["title"]
  d << song["permalink_url"]
  d << song["artwork_url"]
end

CSV.open("./music_data.csv", "wb") do |csv|
  csv << header
  new_data.each {|d| csv << d }
end
