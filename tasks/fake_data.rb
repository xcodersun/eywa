hotels = ["four seasons", "trade winds", "marriott"]
room_types = ["suite", "single", "economy", "luxury"]
brightnesses = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0]
time_start = Time.now.utc - 1.week
time_end = Time.now.utc

hotels.each do |hotel|
  (1..3).to_a.each do |building_id|
    (1..3).to_a.each do |floor_id|
      (1..10).to_a.each do |room_id|
        building = "building-#{building_id}"
        floor = "floor-#{floor_id}"
        room = "room-#{room_id}"
        room_type = room_types[room_id%4]
        device_id = "#{hotel}-#{building}-#{floor}-#{room}"
        url = "ws://159.203.195.73:8081/ws/channels/MQ==/devices/#{device_id}?hotel=#{hotel}&building=#{building}&floor=#{floor}&room=#{room}&room_type=#{room_type}"
        url =  URI.escape url
        ws = WebSocket::Client::Simple.connect url, {headers: {"AccessToken"=>"abcdefg"}}
        t = time_start

        while t <= time_end do
          brightness = brightnesses[t.hour] + rand(3) - 1
          temperature = brightness > 9 ? (80 + rand(10) - 5 + rand) : (65 + rand(6) - 3 + rand)
          cooler = temperature > 73 ? true : false
          sound = cooler ? (80 + rand(10) - 5 ) : (40 + rand(6) - 3)
          light1 = brightness < 4 && temperature < 75
          light2 = brightness >= 4 && brightness < 10
          message = {
            brightness: brightness,
            co2: 70 + rand(5),
            cooler: cooler,
            humidity: 50+rand(10),
            light1: light1,
            light2: light2,
            sound: sound,
            pm2_5: 15 + rand(5),
            temperature: temperature,
            timestamp: t.to_i * 1000
          }
          line = "1|#{t.to_i * 1000 + rand(1000)}|#{message.to_json}"
          puts line
          ws.send(line)
          t = t + 5.minutes
          sleep(0.1)
        end

        ws.close
      end
    end
  end
end
