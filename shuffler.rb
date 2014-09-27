require 'csv'
require 'json'

data = CSV.read("./us-ag-productivity-2004.csv")

header = data.shift

ids = [151370911, 97542154, 104449478, 7910262]

new_data = []
data.each do |d|
  num = rand(10) + 1
  num.times do
    # give them some random ids
    a = [d[0], ids[(rand(10) + 1) % 4]]
    new_data << a
  end
end

# mix it all up
new_data.shuffle!

CSV.open("./state_seed.csv", "wb") do |csv|
  csv << ["state", "track_id"]
  new_data.each {|d| csv << d }
end
