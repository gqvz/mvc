package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", ":3000")
	os.Setenv("JWT_SECRET", "mysecret")
	os.Setenv("DB_DATABASE", "dbname")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "3306")
	defer os.Unsetenv("SERVER_ADDRESS")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DB_DATABASE")
	defer os.Unsetenv("DB_USER")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if cfg.JwtSecret != "mysecret" {
		t.Errorf("Expected JwtSecret to be 'mysecret', got '%s'", cfg.JwtSecret)
	}
	if cfg.ServerAddress != ":3000" {
		t.Errorf("Expected ServerAddress to be ':3000', got '%s'", cfg.ServerAddress)
	}
	if cfg.DB.Database != "dbname" {
		t.Errorf("Expected DB.Database to be 'dbname', got '%s'", cfg.DB.Database)
	}
	if cfg.DB.User != "user" {
		t.Errorf("Expected DB.User to be 'user', got '%s'", cfg.DB.User)
	}
	if cfg.DB.Password != "pass" {
		t.Errorf("Expected DB.Password to be 'pass', got '%s'", cfg.DB.Password)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("Expected DB.Host to be 'localhost', got '%s'", cfg.DB.Host)
	}
	if cfg.DB.Port != 3306 {
		t.Errorf("Expected DB.Port to be 3306, got %d", cfg.DB.Port)
	}
}
