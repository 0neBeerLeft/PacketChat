package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	serverPort  = 9000
	maxMessages = 1000
)

type Message struct {
	Text  string
	Color tcell.Color
}

type ifaceEntry struct {
	NetIface net.Interface
	IP       net.IP
	PcapName string
}

var (
	username string

	screen     tcell.Screen
	screenLock sync.Mutex

	messages []Message
	msgLock  sync.Mutex

	input     string
	inputLock sync.Mutex

	conn         *net.UDPConn
	localSrcPort layers.UDPPort

	deviceName  string
	deviceIP    net.IP
	pcapDevice  string
	broadcastIP string
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: packetchat <username>")
	}
	username = os.Args[1]

	entry, err := selectInterface()
	if err != nil {
		log.Fatal(err)
	}

	deviceName = entry.NetIface.Name
	deviceIP = entry.IP
	pcapDevice = entry.PcapName
	broadcastIP = calcBroadcast(deviceIP)

	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		screen.Fini()
		os.Exit(0)
	}()

	initConnection()

	go sniffLoop()
	go eventLoop()

	addMessage("LAN UDP CHAT", tcell.ColorDodgerBlue)
	addMessage(fmt.Sprintf("User: %s", username), tcell.ColorDodgerBlue)
	addMessage(fmt.Sprintf("Interface: %s", deviceName), tcell.ColorDodgerBlue)
	addMessage(fmt.Sprintf("Broadcast: %s  Port: %d", broadcastIP, serverPort), tcell.ColorDodgerBlue)
	addMessage("Commands: /help  /exit", tcell.ColorDodgerBlue)

	drawScreen()
	select {}
}

func selectInterface() (*ifaceEntry, error) {
	netIfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	pcapDevs, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	var entries []ifaceEntry

	for _, iface := range netIfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil {
				continue
			}
			for _, dev := range pcapDevs {
				for _, paddr := range dev.Addresses {
					if paddr.IP != nil && paddr.IP.Equal(ipnet.IP) {
						entries = append(entries, ifaceEntry{
							NetIface: iface,
							IP:       ipnet.IP,
							PcapName: dev.Name,
						})
					}
				}
			}
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no usable interfaces found")
	}

	fmt.Println("Select interface:")
	for i, e := range entries {
		fmt.Printf("[%d] %s  %s\n", i, e.NetIface.Name, e.IP)
	}

	fmt.Print("Enter number: ")
	var idx int
	fmt.Scan(&idx)

	if idx < 0 || idx >= len(entries) {
		return nil, fmt.Errorf("invalid selection")
	}
	return &entries[idx], nil
}

func calcBroadcast(ip net.IP) string {
	ip4 := ip.To4()
	return fmt.Sprintf("%d.%d.%d.255", ip4[0], ip4[1], ip4[2])
}

func initConnection() {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(broadcastIP),
		Port: serverPort,
	}
	var err error
	conn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	localSrcPort = layers.UDPPort(conn.LocalAddr().(*net.UDPAddr).Port)

	send(fmt.Sprintf("[%s] is online", username), false)
}

func sniffLoop() {
	handle, err := pcap.OpenLive(pcapDevice, 1600, true, pcap.BlockForever)
	if err != nil {
		addMessage("pcap error: "+err.Error(), tcell.ColorRed)
		return
	}
	defer handle.Close()

	handle.SetBPFFilter(fmt.Sprintf("udp and port %d", serverPort))

	src := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range src.Packets() {
		udpLayer := pkt.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			continue
		}

		udp := udpLayer.(*layers.UDP)

		if udp.SrcPort == localSrcPort {
			continue
		}

		raw := strings.TrimSpace(string(bytes.Trim(udp.Payload, "\x00")))
		if raw == "" {
			continue
		}

		addMessage(raw, tcell.ColorYellow)
		drawScreen()
	}
}

func send(text string, user bool) {
	var msg string
	if user {
		msg = fmt.Sprintf("%s: %s", username, text)
	} else {
		msg = text
	}

	if conn == nil {
		addMessage("Not connected", tcell.ColorRed)
		return
	}

	if _, err := conn.Write([]byte(msg)); err != nil {
		addMessage("Send error: "+err.Error(), tcell.ColorRed)
		return
	}

	if user {
		addMessage(msg, tcell.ColorGreen)
	}
}

func addMessage(text string, color tcell.Color) {
	msgLock.Lock()
	defer msgLock.Unlock()
	messages = append(messages, Message{text, color})
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
}

func drawScreen() {
	screenLock.Lock()
	defer screenLock.Unlock()

	screen.Clear()
	w, h := screen.Size()
	drawHeader(w)
	drawMessages(w, h)
	drawInput(w, h)
	screen.Show()
}

func drawHeader(w int) {
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorDodgerBlue)
	for x := 0; x < w; x++ {
		screen.SetContent(x, 0, ' ', nil, style)
	}
	title := " LAN UDP CHAT "
	for i, r := range title {
		screen.SetContent((w-len(title))/2+i, 0, r, nil, style)
	}
}

func drawMessages(w, h int) {
	msgLock.Lock()
	defer msgLock.Unlock()

	y := 1
	total := 0
	for _, m := range messages {
		total += len(wrap(m.Text, w))
	}

	skip := 0
	if total > h-2 {
		skip = total - (h - 2)
	}

	line := 0
	for _, m := range messages {
		for _, l := range wrap(m.Text, w) {
			if line < skip {
				line++
				continue
			}
			printLine(0, y, l, m.Color)
			y++
		}
	}
}

func drawInput(w, h int) {
	inputLock.Lock()
	defer inputLock.Unlock()

	style := tcell.StyleDefault.Foreground(tcell.ColorAqua)
	bar := "> " + input
	if len(bar) > w {
		bar = bar[len(bar)-w:]
	}
	for x := 0; x < w; x++ {
		screen.SetContent(x, h-1, ' ', nil, style)
	}
	for i, r := range bar {
		screen.SetContent(i, h-1, r, nil, style)
	}
}

func printLine(x, y int, text string, color tcell.Color) {
	style := tcell.StyleDefault.Foreground(color)
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}

func wrap(s string, w int) []string {
	var out []string
	for len(s) > w {
		out = append(out, s[:w])
		s = s[w:]
	}
	out = append(out, s)
	return out
}

func eventLoop() {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			drawScreen()
		case *tcell.EventKey:
			inputLock.Lock()
			switch ev.Key() {
			case tcell.KeyEsc:
				screen.Fini()
				os.Exit(0)
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			case tcell.KeyEnter:
				txt := strings.TrimSpace(input)
				if txt != "" {
					if strings.HasPrefix(txt, "/") {
						handleCmd(txt)
					} else {
						send(txt, true)
					}
				}
				input = ""
			default:
				if ev.Rune() != 0 {
					input += string(ev.Rune())
				}
			}
			inputLock.Unlock()
			drawScreen()
		}
	}
}

func handleCmd(cmd string) {
	switch cmd {
	case "/exit", "/e":
		screen.Fini()
		os.Exit(0)
	case "/help", "/h", "/?", "/commands":
		addMessage("COMMAND       DESCRIPTION", tcell.ColorWhite)
		addMessage("/exit, /e     Quit the session", tcell.ColorGray)
		addMessage("/help, /h     Show this list!", tcell.ColorGray)
	default:
		addMessage("Unknown command: "+cmd, tcell.ColorRed)
	}
}
