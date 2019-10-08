package main

import (
	"fmt"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/flv"
	"github.com/nareix/joy4/format/rtmp"
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
		f, _ := os.Create("rtmp_out.flv")

		flvMuxer := flv.NewMuxer(f)

		avutil.CopyFile(flvMuxer, conn)

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
	fmt.Println("rtmp_recorder: server: client closed err:", err)

	return
}
