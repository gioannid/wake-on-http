package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	defaultBroadcast = "255.255.255.255"
	defaultPort      = "8009"
)

var broadcastAddr string
var listenPort string
var defaultMAC string

func wakeDevice(mac string) error {
	packet, err := buildMagicPacket(mac)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", broadcastAddr, 9) // use UDP port 9 (discard) for WOL
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve udp addr: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write(packet); err != nil {
		return fmt.Errorf("failed to send magic packet: %w", err)
	}

	log.Printf("Magic packet sent to %s via %s", mac, addr)
	return nil
}

func buildMagicPacket(mac string) ([]byte, error) {
	clean := strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", "")
	if len(clean) != 12 {
		return nil, fmt.Errorf("invalid MAC length")
	}
	macBytes, err := hex.DecodeString(clean)
	if err != nil || len(macBytes) != 6 {
		return nil, fmt.Errorf("invalid MAC format: %w", err)
	}

	packet := make([]byte, 6+16*6)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], macBytes)
	}
	return packet, nil
}

func isValidMAC(mac string) bool {
	// Check for valid MAC address format (XX:XX:XX:XX:XX:XX)
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	return macRegex.MatchString(mac)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if defaultMAC == "" {
		http.Error(w, "DEFAULT_MAC environment variable not set", http.StatusInternalServerError)
		log.Println("Error: DEFAULT_MAC environment variable not set")
		return
	}

	if !isValidMAC(defaultMAC) {
		http.Error(w, "Invalid DEFAULT_MAC format", http.StatusBadRequest)
		log.Printf("Error: Invalid DEFAULT_MAC format: %s", defaultMAC)
		return
	}

	if err := wakeDevice(defaultMAC); err != nil {
		http.Error(w, fmt.Sprintf("Failed to wake device: %v", err), http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Wake-on-LAN signal sent to %s\n", defaultMAC)
}

func macHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ignore favicon requests
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}

	// Extract MAC from URL path (/MAC_ADDRESS)
	mac := strings.TrimPrefix(r.URL.Path, "/")
	mac = strings.TrimSpace(mac)

	if mac == "" {
		http.Error(w, "MAC address not provided", http.StatusBadRequest)
		return
	}

	if !isValidMAC(mac) {
		http.Error(w, "Invalid MAC address format. Expected: XX:XX:XX:XX:XX:XX", http.StatusBadRequest)
		log.Printf("Error: Invalid MAC format: %s", mac)
		return
	}

	if err := wakeDevice(mac); err != nil {
		http.Error(w, fmt.Sprintf("Failed to wake device: %v", err), http.StatusInternalServerError)
		log.Printf("Error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Wake-on-LAN signal sent to %s\n", mac)
}

func main() {
	// pick broadcast address from env or use default
	broadcastAddr = os.Getenv("BROADCAST_ADDR")
	if broadcastAddr == "" {
		broadcastAddr = defaultBroadcast
	}

	// pick listen port from env or use default
	listenPort = os.Getenv("PORT")
	if listenPort == "" {
		listenPort = defaultPort
	}

	// pick default MAC from env or use empty
	defaultMAC = os.Getenv("DEFAULT_MAC")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			rootHandler(w, r)
		} else {
			macHandler(w, r)
		}
	})

	log.Printf("Starting Wake-on-LAN HTTP server on port %s (broadcast: %s, default MAC: %s)", listenPort, broadcastAddr, defaultMAC)
	if err := http.ListenAndServe(":"+listenPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
