package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mdaffin/go-telegraf"
	ts3 "github.com/thannaske/go-ts3"
)

func main() {

	serverPtr := flag.String("server", "127.0.0.1:10011", "IP address or hostname of Teamspeak 3 server query (default: 127.0.0.1)")
	usernamePtr := flag.String("username", "serveradmin", "Teamspeak 3 server query user (default: serveradmin)")
	passwordPtr := flag.String("password", "", "Teamspeak 3 server query password (default: none)")
	output := flag.String("output", "unix:/var/run/telegraf/telegraf.sock", "Output, can be unix: tcp: or udp:")
	flag.Parse()

	telegrafClient := createClient(output)
	defer telegrafClient.Close()

	tsClient, err := ts3.NewClient(*serverPtr)
	if err != nil {
		logConsole("[Error] Could not establish server query connection to", *serverPtr)
		os.Exit(1)
	}

	defer tsClient.Close()

	if err := tsClient.Login(*usernamePtr, *passwordPtr); err != nil {
		logConsole("[Error] Authentication failure:", err)
		os.Exit(1)
	}

	for range time.Tick(time.Second * 10) {
		if serverList, err := ListServers(tsClient); err != nil {
			logConsole("[Error] Could not iterate through Teamspeak 3 server instances:", err)
			os.Exit(1)
		} else {
			for _, server := range serverList {
				measurement := createMeasurement(server)
				err = telegrafClient.Write(measurement)
				if err != nil {
					logConsole("[Error] Writing measurement failed:", err)
					os.Exit(1)
				}
			}
		}
	}
}

func logConsole(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a)
}

func createClient(url *string) telegraf.Client {
	var err error = nil
	var client telegraf.Client
	if strings.HasPrefix(*url, "unix:") {
		client, err = telegraf.NewUnix(strings.TrimPrefix(*url, "unix:"))
	} else if strings.HasPrefix(*url, "tcp:") {
		client, err = telegraf.NewTCP(strings.TrimPrefix(*url, "tcp:"))
	} else if strings.HasPrefix(*url, "udp:") {
		client, err = telegraf.NewUDP(strings.TrimPrefix(*url, "udp:"))
	}
	if err != nil {
		log.Fatal("could not connect to telegraf:", err)
	}
	return client
}

func createMeasurement(server *ts3.Server) telegraf.Measurement {
	var voiceClients = server.ClientsOnline - server.QueryClientsOnline

	measurement := telegraf.NewMeasurement("teamspeak_server")
	measurement.AddTag("port", strconv.Itoa(server.Port))
	measurement.AddTag("id", strconv.Itoa(server.ID))
	measurement.AddTag("name", server.Name)

	measurement.AddUInt16("port", uint16(server.Port))
	measurement.AddUInt16("id", uint16(server.ID))
	measurement.AddBool("online", server.Status == "online")
	measurement.AddUInt("v_clients", uint(voiceClients))
	measurement.AddUInt("q_clients", uint(server.QueryClientsOnline))
	measurement.AddUInt("m_clients", uint(server.MaxClients))
	measurement.AddBool("autostart", server.AutoStart)
	measurement.AddUInt64("bytes_out", server.BytesSentTotal)
	measurement.AddUInt64("bytes_in", server.BytesReceivedTotal)
	measurement.AddUInt("channels", uint(server.ChannelsOnline))
	measurement.AddUInt("reserved_slots", uint(server.ReservedSlots))
	measurement.AddUInt("uptime", uint(server.Uptime))
	measurement.AddUInt64("packets_in", server.PacketsReceivedTotal)
	measurement.AddUInt64("packets_out", server.PacketsSentTotal)
	measurement.AddInt64("ft_bytes_in_total", server.TotalBytesUploaded)
	measurement.AddInt64("ft_bytes_out_total", server.TotalBytesDownloaded)
	measurement.AddFloat64("pl_control", server.TotalPacketLossControl)
	measurement.AddFloat64("pl_speech", server.TotalPacketLossSpeech)
	measurement.AddFloat64("pl_keepalive", server.TotalPacketLossKeepalive)
	measurement.AddFloat64("pl_total", server.TotalPacketLossTotal)

	measurement.AddUInt64("bytes_out_speech", server.SpeechBytesSent)
	measurement.AddUInt64("bytes_in_speech", server.SpeechBytesReceived)

	measurement.AddUInt64("bytes_out_control", server.ControlBytesSent)
	measurement.AddUInt64("bytes_in_control", server.ControlBytesReceived)

	measurement.AddUInt64("bytes_out_keepalive", server.KeepaliveBytesSent)
	measurement.AddUInt64("bytes_in_keepalive", server.KeepaliveBytesReceived)

	measurement.AddUInt64("packets_out_speech", server.SpeechPacketsSent)
	measurement.AddUInt64("packets_in_speech", server.SpeechPacketsReceived)

	measurement.AddUInt64("packets_out_control", server.ControlPacketsSent)
	measurement.AddUInt64("packets_in_control", server.ControlPacketsReceived)

	measurement.AddUInt64("packets_keepalive_out", server.KeepalivePacketsSent)
	measurement.AddUInt64("packets_keepalive_in", server.KeepalivePacketsReceived)

	measurement.AddFloat32("avg_ping", server.TotalPing)

	return measurement
}

// ListServers iterates through virtual servers and lists extended information about each of them.
func ListServers(client *ts3.Client, options ...string) ([]*ts3.Server, error) {
	var servers []*ts3.Server
	var outputServers []*ts3.Server
	var err error = nil

	if _, err := client.Server.ExecCmd(ts3.NewCmd("serverlist").WithOptions(options...).WithResponse(&servers)); err != nil {
		return nil, err
	}

	info, err := client.Server.Whoami()
	if err != nil {
		return nil, err
	}

	defer func(err *error) {
		if err2 := client.Server.UsePort(info.SelectedServerPort); err2 != nil {
			*err = err2
		}
	}(&err)

	for _, server := range servers {
		if server.Status == "online" {
			if err := client.Server.Use(server.ID); err != nil {
				return nil, err
			}
			if _, err := client.Server.ExecCmd(ts3.NewCmd("serverinfo").WithResponse(&outputServers)); err != nil {
				return nil, err
			}
		} else {
			outputServers = append(outputServers, server)
		}

	}

	return outputServers, err
}
