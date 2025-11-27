// internal/database/database.go
package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

var DB *Database

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "postgres"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "otobo_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Connect() (*Database, error) {
	config := NewConfig()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode,
	)

	// Конфигурация логгера для GORM
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка пула соединений
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = &Database{gormDB}

	log.Println("✅ Database connected successfully")
	return DB, nil
}

// HealthCheck проверяет соединение с базой данных
func (db *Database) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close закрывает соединение с базой данных
func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithTransaction выполняет операции в транзакции
func (db *Database) WithTransaction(fn func(tx *gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// Admin functions for database management

// GetTables возвращает список таблиц в базе данных
func (db *Database) GetTables() ([]string, error) {
	var tables []string
	err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables).Error
	return tables, err
}

// GetTableInfo возвращает информацию о таблице
func (db *Database) GetTableInfo(tableName string) (map[string]interface{}, error) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"name":  tableName,
		"count": count,
	}

	return info, nil
}

// GetTableData возвращает данные из таблицы
func (db *Database) GetTableData(tableName string, limit, offset int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tableName, limit, offset)
	err := db.Raw(query).Scan(&result).Error
	return result, err
}

// GetTableColumns возвращает информацию о колонках таблицы
func (db *Database) GetTableColumns(tableName string) ([]map[string]interface{}, error) {
	var columns []map[string]interface{}
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = ? AND table_schema = 'public'
		ORDER BY ordinal_position
	`
	err := db.Raw(query, tableName).Scan(&columns).Error
	return columns, err
}

// UpdateTableData обновляет данные в таблице
func (db *Database) UpdateTableData(tableName string, id string, data map[string]interface{}) error {
	// Удаляем системные поля если они есть
	delete(data, "id")
	delete(data, "created_at")
	delete(data, "updated_at")

	// Формируем запрос на обновление
	query := fmt.Sprintf("UPDATE %s SET ", tableName)
	values := make([]interface{}, 0)
	setParts := make([]string, 0)

	for key, value := range data {
		setParts = append(setParts, key+" = ?")
		values = append(values, value)
	}

	query += strings.Join(setParts, ", ")
	query += " WHERE id = ?"
	values = append(values, id)

	return db.Exec(query, values...).Error
}

// DeleteTable удаляет таблицу из базы данных
func (db *Database) DeleteTable(tableName string) error {
	return db.Migrator().DropTable(tableName)
}

// BackupDatabase создает резервную копию базы данных
func (db *Database) BackupDatabase() error {
	// В реальной реализации здесь будет код для создания резервной копии
	// Пока просто логируем действие
	log.Println("Creating database backup")
	return nil
}

// OptimizeDatabase оптимизирует базу данных
func (db *Database) OptimizeDatabase() error {
	// В реальной реализации здесь будет код для оптимизации базы данных
	// Пока просто логируем действие
	log.Println("Optimizing database")
	return nil
}

// ClearQueryCache очищает кэш запросов
func (db *Database) ClearQueryCache() error {
	// В реальной реализации здесь будет код для очистки кэша запросов
	// Пока просто логируем действие
	log.Println("Clearing query cache")
	return nil
}
