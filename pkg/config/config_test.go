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
	os.Setenv("DEFAULT_USER_NAME", "admin")
	os.Setenv("DEFAULT_USER_PASSWORD", "admin")
	os.Setenv("DEFAULT_USER_EMAIL", "admin@admin.com")
	defer os.Unsetenv("SERVER_ADDRESS")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DB_DATABASE")
	defer os.Unsetenv("DB_USER")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")
	defer os.Unsetenv("DEFAULT_USER_NAME")
	defer os.Unsetenv("DEFAULT_USER_PASSWORD")
	defer os.Unsetenv("DEFAULT_USER_EMAIL")

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
	if cfg.DB.DefaultUser.Name != "admin" {
		t.Errorf("Expected DefaultUser.Name to be 'admin', got '%s'", cfg.DB.DefaultUser.Name)
	}
	if cfg.DB.DefaultUser.Password != "admin" {
		t.Errorf("Expected DefaultUser.Password to be 'admin', got '%s'", cfg.DB.DefaultUser.Password)
	}
	if cfg.DB.DefaultUser.Email != "admin@admin.com" {
		t.Errorf("Expected DefaultUser.Email to be 'admin@admin.com', got '%s'", cfg.DB.DefaultUser.Email)
	}
}
