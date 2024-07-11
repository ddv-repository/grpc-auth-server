package auth

import "errors"

// https://tools.ietf.org/html/rfc6749#section-5.2
var (
	ErrInvalidRedirectURI   = errors.New("invalid redirect uri")
	ErrInvalidAuthorizeCode = errors.New("invalid authorize code")
	ErrInvalidAccessToken   = errors.New("invalid access token")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrExpiredAccessToken   = errors.New("expired access token")
	ErrExpiredRefreshToken  = errors.New("expired refresh token")

	ErrInvalidRequest          = errors.New("invalid_request")
	ErrUnauthorizedClient      = errors.New("unauthorized_client")
	ErrAccessDenied            = errors.New("access_denied")
	ErrUnsupportedResponseType = errors.New("unsupported_response_type")
	ErrInvalidScope            = errors.New("invalid_scope")
	ErrServerError             = errors.New("server_error")
	ErrTemporarilyUnavailable  = errors.New("temporarily_unavailable")
	ErrInvalidClient           = errors.New("invalid_client")
	ErrInvalidGrant            = errors.New("invalid_grant")
	ErrUnsupportedGrantType    = errors.New("unsupported_grant_type")
)

const (
	ERR_DATABASE_EXCEPTION = "ERR_DATABASE_EXCEPTION"
	ERR_DATABASE_CONN      = "ERR_DATABASE_CONN"
	ERR_AUTH_REQUEST       = "ERR_AUTH_REQUEST"
	ERR_AUTH_FORBIDDEN     = "ERR_AUTH_FORBIDDEN"
	ERR_AUTH_UNAUTHORIZE   = "ERR_AUTH_UNAUTHORIZE"
	ERR_AUTH_EXCEPTION     = "ERR_AUTH_EXCEPTION"
)
