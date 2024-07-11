package main

import (
	"fmt"
	"net"
	"os"

	"grpc-auth-server/cmd/config"
	"grpc-auth-server/internal/auth"
	"grpc-auth-server/internal/server"
	pb "grpc-auth-server/protogen/token"

	hclog "github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	logger hclog.Logger
	rdx    *auth.PoolRedisClient
)

func init() {
	logger = hclog.NewInterceptLogger(&hclog.LoggerOptions{})

	config.LoadConfig()

	var err error
	rdx, err = auth.NewPoolRedisClient(fmt.Sprintf("%s:%s", config.ConfigData.RedisHost, config.ConfigData.RedisPort))
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	clientID := config.ConfigData.ClientID
	clientSecret := config.ConfigData.ClientSecret

	key := fmt.Sprintf("log_access:%s", clientID)
	if _, err := rdx.Get(key); err != nil {
		token := auth.AccessToken{
			ClientID:  clientID,
			UserID:    "admin",
			ExpiredIn: 0,
			OneTime:   false,
			Scope:     []string{"all"},
			IPAddress: []string{},
		}

		serializedToken, err := auth.Serialize(token)
		if err != nil {
			logger.Error("failed to serialize token", "error", err)
			os.Exit(1)
		}

		encryptedToken, err := auth.Encrypt(serializedToken, clientSecret)
		if err != nil {
			logger.Error("failed to encrypt token", "error", err)
			os.Exit(1)
		}

		encryptedTokenString := auth.Base62Encode(encryptedToken)

		rdx.Set(key, encryptedTokenString, 0)
		rdx.HMSet(encryptedTokenString, config.StructToMap(token), 0)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":"+config.ConfigData.ApiPort)
	if err != nil {
		logger.Error("failed to listen", "error", err)
	}

	srv := grpc.NewServer()
	pb.RegisterTokenServer(srv, &server.TokenServer{RedisClient: rdx})

	reflection.Register(srv)

	logger.Info("Server listening", "port", listener.Addr())
	if err := srv.Serve(listener); err != nil {
		logger.Error("dailed to server", "error", err)
	}
}
