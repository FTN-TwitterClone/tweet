package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"os"
)

func GetHTTPServerTLSConfig() *tls.Config {
	var caCert []byte
	var err error
	var caCertPool *x509.CertPool

	caCert, err = ioutil.ReadFile(os.Getenv("CA_CERT"))
	if err != nil {
		log.Fatal("Error opening cert file", err)
	}

	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		ServerName: "auth",
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12, // TLS versions below 1.2 are considered insecure - see https://www.rfc-editor.org/rfc/rfc7525.txt for details
	}
}

func GetgRPCServerTLSConfig() *tls.Config {
	serverCertPath := os.Getenv("CERT")
	serverKeyPath := os.Getenv("KEY")
	caCertPath := os.Getenv("CA_CERT")

	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		log.Fatalf("Failed to load server certificate and key. %s.", err)
	}

	trustedCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Failed to load trusted certificate. %s.", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		log.Fatalf("Failed to append trusted certificate to certificate pool. %s.", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}
}

func GetgRPCClientTLSConfig() *tls.Config {
	clientCertPath := os.Getenv("CERT")
	clientKeyPath := os.Getenv("KEY")
	caCertPath := os.Getenv("CA_CERT")

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		log.Fatalf("Failed to load client certificate and key. %s.", err)
	}

	trustedCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("Failed to load trusted certificate. %s.", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		log.Fatalf("Failed to append trusted certificate to certificate pool. %s.", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}
}

func GetgRPCConnection(address string) (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(GetgRPCClientTLSConfig())

	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("Failed to start gRPC connection: %v", err)
	}

	return conn, err
}
