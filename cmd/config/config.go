package config

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	pb "grpc-auth-server/protogen/token"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		Namespace     string `json:"namespace"`
		Version       string `json:"version"`
		PostgresHost  string `json:"postgres_host"`
		PostgresPort  string `json:"postgres_port"`
		PostgresDB    string `json:"postgres_db"`
		PostgresUser  string `json:"postgres_user"`
		PostgresPass  string `json:"postgres_password"`
		RedisHost     string `json:"redis_host"`
		RedisUsername string `json:"redis_username"`
		RedisPort     string `json:"redis_port"`
		RedisPassword string `json:"redis_password"`
		ClientID      string `json:"client_id"`
		ClientSecret  string `json:"client_secret"`
		ApiPort       string `json:"api_port"`
	}
)

var ConfigData Config

func LoadConfig() {
	godotenv.Load()

	ConfigData.Namespace = os.Getenv("NAMESPACE")
	ConfigData.Version = os.Getenv("VERSION")
	ConfigData.PostgresHost = os.Getenv("POSTGRES_HOST")
	ConfigData.PostgresPort = os.Getenv("POSTGRES_PORT")
	ConfigData.PostgresUser = os.Getenv("POSTGRES_USER")
	ConfigData.PostgresPass = os.Getenv("POSTGRES_PASSWORD")
	ConfigData.PostgresDB = os.Getenv("POSTGRES_DB")

	ConfigData.RedisHost = os.Getenv("REDIS_HOST")
	ConfigData.RedisUsername = os.Getenv("REDIS_USERNAME")
	ConfigData.RedisPort = os.Getenv("REDIS_PORT")
	ConfigData.RedisPassword = os.Getenv("REDIS_PASSWORD")

	ConfigData.ClientID = os.Getenv("AUTH_CLIENT_ID")
	ConfigData.ClientSecret = os.Getenv("AUTH_CLIENT_SECRET")

	ConfigData.ApiPort = os.Getenv("AUTH_SERVER_API__PORT")
}

func StructToMap(obj interface{}) map[string]interface{} {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	objType := objValue.Type()
	result := make(map[string]interface{})

	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i).Interface()
		result[field.Name] = fieldValue
	}

	return result
}

func ValidatonGenerateTokenRequest(param *pb.GenerateTokenRequest) error {
	if param.ClientId == "" {
		return fmt.Errorf("invalid client id parameter: input empty")
	}
	if param.ClientSecret == "" {
		return fmt.Errorf("invalid client secret parameter: input empty")
	}
	if param.Expiry == "" {
		return fmt.Errorf("invalid expiry parameter: input empty")
	}
	return nil
}

func ValidateTokenRequest(param *pb.ValidateTokenRequest) error {
	if param.Key == "" {
		return fmt.Errorf("invalid key: input empty")
	}
	if param.Token == "" {
		return fmt.Errorf("invalid token: input empty")
	}
	return nil
}

func ParseExpiry(expiry string) (time.Duration, error) {
	switch {
	case strings.HasSuffix(expiry, "d"):
		daysStr := strings.TrimSuffix(expiry, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, fmt.Errorf("invalid days value: %v", err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	case strings.HasSuffix(expiry, "w"):
		weeksStr := strings.TrimSuffix(expiry, "w")
		weeks, err := strconv.Atoi(weeksStr)
		if err != nil {
			return 0, fmt.Errorf("invalid weeks value: %v", err)
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	case strings.HasSuffix(expiry, "mo"):
		monthsStr := strings.TrimSuffix(expiry, "mo")
		months, err := strconv.Atoi(monthsStr)
		if err != nil {
			return 0, fmt.Errorf("invalid months value: %v", err)
		}
		return time.Duration(months) * 30 * 24 * time.Hour, nil
	case strings.HasSuffix(expiry, "yr"):
		yearsStr := strings.TrimSuffix(expiry, "yr")
		years, err := strconv.Atoi(yearsStr)
		if err != nil {
			return 0, fmt.Errorf("invalid years value: %v", err)
		}
		return time.Duration(years) * 365 * 24 * time.Hour, nil
	default:
		return time.ParseDuration(expiry)
	}
}

func ValidateDomainOrIPAddress(input string) bool {
	return isValidDomain(input) || isValidIPAddress(input)
}

func isValidDomain(domain string) bool {
	domainRegex := `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
	re := regexp.MustCompile(domainRegex)
	return re.MatchString(domain)
}

func isValidIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}
