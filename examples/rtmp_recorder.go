package main

import (
	"fmt"
	"github.com/nareix/joy4/av"
	//"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/flv"
	"github.com/nareix/joy4/format/rtmp"
	"io"
	"net"
	"os"
)

func init() {
	format.RegisterAll()
}

func main() {

	rtmp.Debug = true
	server := &rtmp.Server{}

	server.HandlePublish = func(conn *rtmp.Conn) {

		fmt.Println("rtmp_recorder: new connection", conn.URL)

		f1, _ := os.Create("rtmp_out_1.flv")
		flvMuxer := flv.NewMuxer(f1)
		CopyFile(flvMuxer, conn)
		f1.Close()

		f2, _ := os.Create("rtmp_out_2.flv")
		flvMuxer2 := flv.NewMuxer(f2)
		CopyFile(flvMuxer2, conn)
		f2.Close()
	}

	// Serve one single connection - for a full server use 'server.ListenAndServe()'
	serveOneConnection(server)

	// ffmpeg -re -i movie.flv -c copy -f flv rtmp://localhost/movie
	// ffmpeg -f avfoundation -i "0:0" .... -f flv rtmp://localhost/screen

}

func serveOneConnection(server *rtmp.Server) (err error) {

	addr := ":1935"

	var tcpaddr *net.TCPAddr
	if tcpaddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
		err = fmt.Errorf("rtmp_recorder: ListenAndServe: %s", err)
		return
	}

	var listener *net.TCPListener
	if listener, err = net.ListenTCP("tcp", tcpaddr); err != nil {
		return
	}

	var netconn net.Conn
	if netconn, err = listener.Accept(); err != nil {
		return
	}

	fmt.Println("rtmp_recorder: server: accepted")

	err = server.Server(netconn)
	return
}

var pktLimit = 60

func CopyPackets(dst av.PacketWriter, src av.PacketReader) (err error) {
	for npkts := 0; npkts < pktLimit; npkts++ {
		var pkt av.Packet
		if pkt, err = src.ReadPacket(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if err = dst.WritePacket(pkt); err != nil {
			return
		}
	}
	fmt.Println("rtmp_recorder: packet limit reached", "limit", pktLimit)
	return
}

func CopyFile(dst av.Muxer, src av.Demuxer) (err error) {
	var streams []av.CodecData
	if streams, err = src.Streams(); err != nil {
		return
	}
	if err = dst.WriteHeader(streams); err != nil {
		return
	}

	if err = CopyPackets(dst, src); err != nil {
		if err != io.EOF {
			return
		}
	}
	if err = dst.WriteTrailer(); err != nil {
		return
	}
	return
}
