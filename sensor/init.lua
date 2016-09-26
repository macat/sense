dofile("config.lua")

wifi_ok = 0
mqtt_ok = 0

wifi.setmode(wifi.STATION)
wifi.sta.config(wifi_ssid,wifi_password)
wifi.sta.autoconnect(1)
print("connected to wifi!")
wifi.sta.autoconnect(1)
print(wifi.sta.getip())
print("hello!")


gpio.mode(6, gpio.INPUT)


tmr.alarm(0, 1000, tmr.ALARM_AUTO, function()
  print("WIFI"..wifi.sta.status())
   if wifi.sta.status() == 5 then
      wifi_ok = 1
      tmr.stop(0)
      sntp.sync('0.north-america.pool.ntp.org', 
        function(sec,usec,server)
          print('time sync', sec, usec, server)
        end,
        function()
          print('time sync failed!')
        end
      )
   else
       wifi_ok = 0
   end
end)


-- Connect to mqtt
tmr.alarm(1, 1000, tmr.ALARM_AUTO, function()
  if wifi_ok == 1 then
    mqtt_connect()
  end
end)

-- Collect temperature/humidity data every minut
tmr.alarm(2, 60000, tmr.ALARM_AUTO, function()
  if mqtt_ok == 1 then
    temperature()
  end
end)

-- Collect motion data every second
tmr.alarm(3, 1000, tmr.ALARM_AUTO, function()
  if mqtt_ok == 1 then
    motion()
  end
end)

-- Restart node ever 15 minutes
tmr.alarm(4, 900000, 1, function()
  node.restart()
end)


function mqtt_connect()
  m = mqtt.Client("MCU"..device_id, 120, mqtt_user, mqtt_password)
  print("connecting")
  m:connect(mqtt_ip , 1883, 0,
  function(conn)
    print("Connected to MQTT")
    sec, usec = rtctime.get()
    m:publish("sense/"..location.."/"..floor.."/"..room.."/"..device_id.."/status",sec..",1",0,0)
    tmr.stop(1)
    mqtt_ok = 1;
  end, function(client, reason)
    print("failed reason: "..reason)
  end)
end

function temperature()
  status, temp, humi, temp_dec, humi_dec = dht.read(5)
  if status == dht.OK then
      print("DHT Temperature:"..temp..";".."Humidity:"..humi)
      sec, usec = rtctime.get()
      m:publish("sense/"..location.."/"..floor.."/"..room.."/"..device_id.."/temp", sec..","..temp, 0, 0)
      m:publish("sense/"..location.."/"..floor.."/"..room.."/"..device_id.."/humid", sec..","..humi, 0,0)
  elseif status == dht.ERROR_CHECKSUM then
      print( "DHT Checksum error." )
  elseif status == dht.ERROR_TIMEOUT then
      print( "DHT timed out." )
  end
end

function motion()
  if gpio.read(6) == 1 then
    sec, usec = rtctime.get()
    print(sec.."movement")
    m:publish("sense/"..location.."/"..floor.."/"..room.."/"..device_id.."/motion", sec..",1", 0,0)
  end
end

