package main

import (
	"fmt"
	"github.com/nareix/joy4/av"
	//"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/format"
	"github.com/nareix/joy4/format/ts"
	"io"
	"net"
	"os"
)

const frameRate = 24
const pktLimit = frameRate * 30
const dbg = false

func init() {
	format.RegisterAll()
}

func main() {

	serveOneConnection()

}

func record(demux av.Demuxer, i int) error {

	f1, err := os.Create(fmt.Sprintf("ts_out_%00d.ts", i))
	if err != nil {
		fmt.Println("ts_recorder: failed to open recording", "err", err)
		return err
	}
	mux := ts.NewMuxer(f1)

	CopyFile(mux, demux)
	f1.Close()

	return nil
}

func readUdp(pc net.PacketConn, w io.Writer) error {

	buf := make([]byte, 65536)

	for {
		n, sender, err := pc.ReadFrom(buf)
		if err != nil {
			fmt.Println("ts_recorder: UDP read failed", "err", err)
			return err
		}

		if dbg {
			fmt.Println("te_recorder: packet received", "bytes", n, "from", sender.String())
		}

		bw, err := w.Write(buf[:n])
		if err != nil || bw != n {
			fmt.Println("ts_recorder: write failed", "err", err, "bw/n", bw, n)
			return err
		}
	}
	return nil
}

func serveOneConnection() (err error) {

	pr, pw := io.Pipe()

	addr := ":12001"

	pc, err := net.ListenPacket("udp", addr)
	if err != nil {
		return
	}

	fmt.Println("ts_recorder: server: accepted")

	demux := ts.NewDemuxer(pr)

	done := make(chan bool)

	go func() {
		readUdp(pc, pw)
	}()

	go func() {
		for i := 1; i < 4; i++ {
			err := record(demux, i)
			if err != nil {
				fmt.Println("rtmp_recorder: recording failed", "err", err)
			}
		}
		done <- true
	}()

	<-done

	return
}

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
		fmt.Println("rtmp_recorder: copy packet", "npkts", npkts)
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
