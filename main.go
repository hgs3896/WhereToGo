package main

import (
	"fmt"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

/* ListenDNSPacket is a function to listening for DNS packets from local DNS resolver(client) */
func ListenDNSPacket(conn **net.UDPConn) (err error) {
	*conn, err = net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	return err
}

func ReadPacket(conn *net.UDPConn) (error, []byte) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return err, []byte{}
	}
	fmt.Printf("%d bytes received from DNS resolver.\n\n", n)
	return err, buf[:n]
}

func WritePacket(conn *net.UDPConn, packet []byte) error {
	_, err := conn.Write(packet)
	return err
}

func PrintDNSMessage(msg dnsmessage.Message) {
	fmt.Printf("DNS Header: %s\n", msg.Header.GoString())
	for idx, q := range msg.Questions {
		fmt.Printf("[Q%02d] %s\n", idx+1, q.GoString())
	}
	for idx, a := range msg.Answers {
		fmt.Printf("[A%02d] %s\n", idx+1, a.GoString())
	}
}

func ProcessDNSPacket(packet []byte) (error, []byte) {
	var msg dnsmessage.Message
	var responseMsg dnsmessage.Message

	if err := msg.Unpack(packet); err != nil {
		return err, []byte{}
	}

	PrintDNSMessage(msg)

	buf := make([]byte, 2, 514)
	b := dnsmessage.NewBuilder(buf, msg.Header)
	b.Question(msg.Questions[0])
	b.AResource(
		dnsmessage.ResourceHeader{
			Name:  msg.Questions[0].Name,
			Type:  msg.Questions[0].Type,
			Class: msg.Questions[0].Class,
			TTL:   148,
		}, dnsmessage.AResource{A: [...]byte{127, 0, 0, 1}},
	)
	buf, err := b.Finish()
	responseMsg.Unpack(buf)
	PrintDNSMessage(responseMsg)

	return err, buf
}

func main() {
	var conn *net.UDPConn
	var err error
	var packet []byte

	if err = ListenDNSPacket(&conn); err != nil {
		fmt.Printf(err.Error())
		return
	}

	for {
		if err, packet = ReadPacket(conn); err != nil {
			fmt.Println(err)
			break
		}

		err, response := ProcessDNSPacket(packet)

		if err != nil {
			fmt.Println(err)
			break
		}

		if err = WritePacket(conn, response); err != nil {
			fmt.Println(err)
			break
		}
	}

	fmt.Println("Finished")
}
