package main

import (
	"flag"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	proxyAddrStr  string
	serverAddrStr string
)

type session struct {
	serverConn *net.UDPConn
	clientAddr *net.UDPAddr
	lastActive time.Time
	closeChan  chan struct{}
	mu         sync.Mutex
}

func init() {
	flag.StringVar(&proxyAddrStr, "proxy", "", "Proxy address (e.g., :9987). Can also be set via PROXY_ADDR environment variable.")
	flag.StringVar(&proxyAddrStr, "p", "", "Proxy address (shorthand)")
	flag.StringVar(&serverAddrStr, "server", "", "TeamSpeak 3 server address (e.g., ts3.example.com:9987). Can also be set via SERVER_ADDR environment variable.")
	flag.StringVar(&serverAddrStr, "s", "", "TeamSpeak 3 server address (shorthand)")
}

func main() {
	flag.Parse()

	// Handle proxy address configuration
	if proxyAddrStr == "" {
		proxyAddrStr = os.Getenv("PROXY_ADDR")
		if proxyAddrStr == "" {
			proxyAddrStr = ":9987" // Default proxy address
		}
	}

	// Handle server address configuration
	if serverAddrStr == "" {
		serverAddrStr = os.Getenv("SERVER_ADDR")
		if serverAddrStr == "" {
			log.Fatal("Server address must be provided via -server flag or SERVER_ADDR environment variable")
		}
	}

	proxyAddr, err := net.ResolveUDPAddr("udp", proxyAddrStr)
	if err != nil {
		log.Fatalf("Failed to resolve proxy address %q: %v", proxyAddrStr, err)
	}

	proxyConn, err := net.ListenUDP("udp", proxyAddr)
	if err != nil {
		log.Fatal("Failed to start proxy listener:", err)
	}
	defer proxyConn.Close()

	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		log.Fatalf("Failed to resolve server address %q: %v", serverAddrStr, err)
	}

	var sessions sync.Map
	go cleanupSessions(&sessions)

	buffer := make([]byte, 1500) // MTU size
	for {
		n, clientAddr, err := proxyConn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from client: %v", err)
			continue
		}

		data := make([]byte, n)
		copy(data, buffer[:n])

		go handleClientPacket(data, clientAddr, proxyConn, &sessions, serverAddr)
	}
}

func handleClientPacket(data []byte, clientAddr *net.UDPAddr, proxyConn *net.UDPConn, sessions *sync.Map, serverAddr *net.UDPAddr) {
	key := clientAddr.String()

	rawSession, exists := sessions.Load(key)
	if !exists {
		serverConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			log.Printf("Failed to create server connection: %v", err)
			return
		}

		s := &session{
			serverConn: serverConn,
			clientAddr: clientAddr,
			lastActive: time.Now(),
			closeChan:  make(chan struct{}),
		}

		sessions.Store(key, s)
		go runSessionServerReader(s, proxyConn, sessions)

		rawSession = s
	}

	s := rawSession.(*session)
	s.mu.Lock()
	s.lastActive = time.Now()
	s.mu.Unlock()

	_, err := s.serverConn.WriteToUDP(data, serverAddr)
	if err != nil {
		log.Printf("Failed to forward packet to server: %v", err)
		sessions.Delete(key)
		s.serverConn.Close()
		close(s.closeChan)
	}
}

func runSessionServerReader(s *session, proxyConn *net.UDPConn, sessions *sync.Map) {
	defer s.serverConn.Close()
	defer close(s.closeChan)

	buffer := make([]byte, 1500)
	for {
		select {
		case <-s.closeChan:
			return
		default:
			s.serverConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, _, err := s.serverConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					s.mu.Lock()
					inactive := time.Since(s.lastActive) > 5*time.Minute
					s.mu.Unlock()
					if inactive {
						sessions.Delete(s.clientAddr.String())
						return
					}
					continue
				}
				log.Printf("Server read error: %v", err)
				sessions.Delete(s.clientAddr.String())
				return
			}

			_, err = proxyConn.WriteToUDP(buffer[:n], s.clientAddr)
			if err != nil {
				log.Printf("Client write error: %v", err)
				sessions.Delete(s.clientAddr.String())
				return
			}

			s.mu.Lock()
			s.lastActive = time.Now()
			s.mu.Unlock()
		}
	}
}

func cleanupSessions(sessions *sync.Map) {
	for range time.Tick(1 * time.Minute) {
		sessions.Range(func(key, value interface{}) bool {
			s := value.(*session)
			s.mu.Lock()
			inactive := time.Since(s.lastActive) > 5*time.Minute
			s.mu.Unlock()

			if inactive {
				sessions.Delete(key)
				close(s.closeChan)
			}
			return true
		})
	}
}
