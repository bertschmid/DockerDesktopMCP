package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"docker-mcp/internal/docker"
	"docker-mcp/internal/mcp"
)

func main() {
	port := flag.Int("port", 3282, "Port to listen on (default: 3282)")
	host := flag.String("host", "127.0.0.1", "Host to bind to (default: 127.0.0.1)")
	tlsEnabled := flag.Bool("tls", false, "Enable TLS/HTTPS")
	certFile := flag.String("cert", "", "TLS certificate file (auto-generated if empty)")
	keyFile := flag.String("key", "", "TLS key file (auto-generated if empty)")
	flag.Parse()

	// Environment variable overrides
	if v := os.Getenv("MCP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			*port = p
		}
	}
	if v := os.Getenv("MCP_HOST"); v != "" {
		*host = v
	}
	if os.Getenv("MCP_TLS") == "true" {
		*tlsEnabled = true
	}
	if v := os.Getenv("MCP_CERT"); v != "" {
		*certFile = v
	}
	if v := os.Getenv("MCP_KEY"); v != "" {
		*keyFile = v
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Verify Docker connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dockerClient.Ping(ctx); err != nil {
		log.Fatalf("Cannot connect to Docker daemon: %v", err)
	}
	log.Println("Connected to Docker daemon")

	// Create MCP server
	mcpHandler := mcp.NewServer(dockerClient)

	// Setup HTTP server
	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	srv := &http.Server{
		Handler:      mcpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	scheme := "http"
	if *tlsEnabled {
		scheme = "https"
		tlsConfig, err := buildTLSConfig(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("TLS configuration failed: %v", err)
		}
		srv.TLSConfig = tlsConfig
		listener = tls.NewListener(listener, tlsConfig)
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("╔════════════════════════════════════════════╗")
		log.Printf("║  Docker Desktop MCP Server v1.0.0          ║")
		log.Printf("╠════════════════════════════════════════════╣")
		log.Printf("║  Endpoint : %s://%s/mcp   ", scheme, addr)
		log.Printf("║  Health   : %s://%s/health", scheme, addr)
		log.Printf("╚════════════════════════════════════════════╝")
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

// buildTLSConfig loads or auto-generates a self-signed TLS certificate
func buildTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	var cert tls.Certificate
	var err error

	if certFile != "" && keyFile != "" {
		cert, err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("loading TLS key pair: %w", err)
		}
		log.Printf("TLS: loaded certificate from %s", certFile)
	} else {
		cert, err = generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("generating self-signed cert: %w", err)
		}
		log.Println("TLS: using auto-generated self-signed certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// generateSelfSignedCert creates a self-signed TLS certificate in memory
func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization:       []string{"Docker Desktop MCP"},
			OrganizationalUnit: []string{"MCP Server"},
		},
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}
