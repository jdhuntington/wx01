// replay sends sample Tempest UDP packets to localhost for testing.
// Usage: go run ./cmd/replay/
package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:50222")
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	now := time.Now().Unix()

	packets := []string{
		// obs_st
		fmt.Sprintf(`{"serial_number":"ST-00000512","type":"obs_st","hub_sn":"HB-00013030","obs":[[%d,0.18,0.22,0.27,144,6,1017.57,22.37,50.26,328,0.03,3,0.000000,0,0,0,2.410,1]],"firmware_revision":129}`, now),
		// rapid_wind
		fmt.Sprintf(`{"serial_number":"ST-00000512","type":"rapid_wind","hub_sn":"HB-00013030","ob":[%d,2.3,128]}`, now),
		// evt_precip
		fmt.Sprintf(`{"serial_number":"ST-00000512","type":"evt_precip","hub_sn":"HB-00013030","evt":[%d]}`, now),
		// evt_strike
		fmt.Sprintf(`{"serial_number":"ST-00000512","type":"evt_strike","hub_sn":"HB-00013030","evt":[%d,27,3848]}`, now),
		// device_status
		fmt.Sprintf(`{"serial_number":"ST-00000512","type":"device_status","hub_sn":"HB-00013030","timestamp":%d,"uptime":2189,"voltage":3.50,"firmware_revision":17,"rssi":-17,"hub_rssi":-87,"sensor_status":0,"debug":0}`, now),
		// hub_status
		fmt.Sprintf(`{"serial_number":"HB-00013030","type":"hub_status","firmware_revision":"35","uptime":1670133,"rssi":-62,"timestamp":%d,"reset_flags":"BOR,PIN,POR","seq":48,"fs":[1,0,15675411,524288],"radio_stats":[2,1,0,3,2839],"mqtt_stats":[1,0]}`, now),
	}

	for _, p := range packets {
		_, err := conn.Write([]byte(p))
		if err != nil {
			fmt.Printf("send error: %v\n", err)
		} else {
			fmt.Printf("sent: %.60s...\n", p)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("done — sent 6 sample packets")
}
