// internal/server/token_server.go
package server

import (
	"context"
	"fmt"

	"grpc-auth-server/cmd/config"
	"grpc-auth-server/internal/auth"
	"grpc-auth-server/pkg/response"
	pb "grpc-auth-server/protogen/token"
)

type (
	TokenServer struct {
		pb.UnimplementedTokenServer
		RedisClient *auth.PoolRedisClient
	}

	TokenResponse struct {
		ClientID  string   `json:"client_id"`
		UserID    string   `json:"user_id,omitempty"`
		Scope     []string `json:"scope"`
		IPAddress []string `json:"ip_address"`
		OneTime   bool     `json:"one_time"`
		TokenType string   `json:"token_type,omitempty"`
	}
)

func (srv *TokenServer) GenerateToken(ctx context.Context, req *pb.GenerateTokenRequest) (*pb.GenerateTokenResponse, error) {
	if err := config.ValidatonGenerateTokenRequest(req); err != nil {
		return &pb.GenerateTokenResponse{
			Error:     auth.ERR_AUTH_REQUEST,
			ErrorCode: int32(response.ErrorLine()),
			Status:    err.Error(),
		}, nil
	}

	duration, err := config.ParseExpiry(req.Expiry)
	if err != nil {
		return &pb.GenerateTokenResponse{
			Error:     auth.ERR_AUTH_REQUEST,
			ErrorCode: int32(response.ErrorLine()),
			Status:    err.Error(),
		}, nil
	}

	if len(req.IpAddress) > 0 {
		for _, ipaddr := range req.IpAddress {
			if !config.ValidateDomainOrIPAddress(ipaddr) {
				return &pb.GenerateTokenResponse{
					Error:     auth.ERR_AUTH_REQUEST,
					ErrorCode: int32(response.ErrorLine()),
					Status:    "invalid ip address filtering: input invalid ip_address",
				}, nil
			}
		}
	}

	key := fmt.Sprintf("log_access:%s", req.ClientId)
	logAccess, errLogAccess := srv.RedisClient.Get(key)

	token := auth.AccessToken{
		ClientID:  req.ClientId,
		UserID:    req.UserId,
		ExpiredIn: duration,
		OneTime:   req.OneTime,
		Scope:     req.Scope,
		IPAddress: req.IpAddress,
	}

	serializedToken, err := auth.Serialize(token)
	if err != nil {
		return &pb.GenerateTokenResponse{
			Error:     auth.ERR_AUTH_REQUEST,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "failed to serialized token",
		}, nil
	}

	encryptedToken, err := auth.Encrypt(serializedToken, req.ClientSecret)
	if err != nil {
		return &pb.GenerateTokenResponse{
			Error:     auth.ERR_AUTH_REQUEST,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "failed to generate token",
		}, nil
	}

	encryptedTokenString := auth.Base62Encode(encryptedToken)

	srv.RedisClient.Set(key, encryptedTokenString, 0)
	srv.RedisClient.HMSet(encryptedTokenString, config.StructToMap(token), 0)

	if errLogAccess == nil {
		srv.RedisClient.Del(logAccess)
	}

	durationSeconds := int(duration.Seconds())
	if duration != 0 {
		srv.RedisClient.Expire(encryptedTokenString, int(durationSeconds))
	}

	return &pb.GenerateTokenResponse{
		Error:     "",
		ErrorCode: 0,
		ExpiredIn: int32(duration),
		Scope:     req.Scope,
		Status:    "ok",
		Token:     encryptedTokenString,
		TokenType: "Bearer",
	}, nil
}

func (srv *TokenServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if err := config.ValidateTokenRequest(req); err != nil {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_REQUEST,
			ErrorCode: int32(response.ErrorLine()),
			Status:    err.Error(),
		}, nil
	}

	valid, err := srv.RedisClient.HGetAll(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_FORBIDDEN,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "unauthorized",
		}, nil
	}

	if len(valid) == 0 {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_EXCEPTION,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "unauthorized",
		}, nil
	}

	decodedEncryptedToken, err := auth.Base62Decode(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_EXCEPTION,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "unauthorized",
		}, nil
	}

	decryptedData, err := auth.Decrypt(decodedEncryptedToken, req.Key)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_EXCEPTION,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "unauthorized",
		}, nil
	}

	deserializedToken, err := auth.Deserialize(decryptedData)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error:     auth.ERR_AUTH_EXCEPTION,
			ErrorCode: int32(response.ErrorLine()),
			Status:    "unauthorized",
		}, nil
	}

	matchIP := false
	if len(deserializedToken.IPAddress) != 0 {
		for _, ip := range deserializedToken.IPAddress {
			if ip == req.ClientIp {
				matchIP = true
			}
		}

		if !matchIP {
			return &pb.ValidateTokenResponse{
				Error:     auth.ERR_AUTH_EXCEPTION,
				ErrorCode: int32(response.ErrorLine()),
				Status:    "unauthorized",
			}, nil
		}
	}

	return &pb.ValidateTokenResponse{
		Data: &pb.TokenData{
			ClientId:  deserializedToken.ClientID,
			UserId:    deserializedToken.UserID,
			Scope:     deserializedToken.Scope,
			IpAddress: deserializedToken.IPAddress,
			OneTime:   deserializedToken.OneTime,
		},
		Error:     "",
		ErrorCode: 0,
		Status:    "ok",
	}, nil
}
