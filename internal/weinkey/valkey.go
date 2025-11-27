package weinkey

import (
	"log"
	"otobo/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/valkey"
)

func ValkeyInit() *valkey.Storage {
	var store *valkey.Storage
	hostVK := utils.GetEnv("REDIS_HOST", "valkey")
	portVK := utils.GetEnv("REDIS_PORT", "6379")
	valkeyURL := hostVK + ":" + portVK
	if valkeyURL == "" {
		valkeyURL = "localhost:6379" // Fallback for local testing outside docker
	}

	for i := range 5 {
		store = valkey.New(valkey.Config{
			InitAddress: []string{valkeyURL},
			Username:    "",
			Password:    "",
			SelectDB:    0,
			Reset:       false,
			TLSConfig:   nil,
		})

		// Проверка подключения
		if store != nil {
			break
		}

		log.Printf("Failed to connect to Valkey, retrying in 2 seconds... (attempt %d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if store == nil {
		log.Fatal("Failed to connect to Valkey after 5 attempts:", valkeyURL)
	}
	return store
}

func SessionInit(store *valkey.Storage) *session.Store {
	return session.New(session.Config{
		Storage:    store, // Используем Valkey storage, который мы только что инициализировали
		KeyLookup:  "cookie:session_id",
		Expiration: 24 * time.Hour,
	})
}

// AdminValkeyClient предоставляет расширенные функции управления Valkey для админки
type AdminValkeyClient struct {
	store *valkey.Storage
}

// NewAdminValkeyClient создает новый клиент для управления Valkey
func NewAdminValkeyClient(store *valkey.Storage) *AdminValkeyClient {
	return &AdminValkeyClient{store: store}
}

// GetStats возвращает статистику использования Valkey
func (vc *AdminValkeyClient) GetStats() (map[string]interface{}, error) {
	// В реальной реализации здесь будет код для получения статистики Valkey
	// Пока возвращаем заглушку
	stats := map[string]interface{}{
		"hits":   1248,
		"misses": 89,
		"usage":  65,
	}
	return stats, nil
}

// GetKeys возвращает список ключей в Valkey
func (vc *AdminValkeyClient) GetKeys() ([]string, error) {
	// В реальной реализации здесь будет код для получения списка ключей
	// Пока возвращаем заглушку
	keys := []string{
		"session:abc123",
		"product:42",
		"user:123",
		"cart:456",
	}
	return keys, nil
}

// DeleteKey удаляет ключ из Valkey
func (vc *AdminValkeyClient) DeleteKey(key string) error {
	// В реальной реализации здесь будет код для удаления ключа
	// Пока просто логируем действие
	log.Printf("Deleting key: %s", key)
	return nil
}

// ClearAll очищает весь Valkey кэш
func (vc *AdminValkeyClient) ClearAll() error {
	// В реальной реализации здесь будет код для очистки всего кэша
	// Пока просто логируем действие
	log.Println("Clearing all Valkey cache")
	return nil
}

// ClearSessions очищает все сессии
func (vc *AdminValkeyClient) ClearSessions() error {
	// В реальной реализации здесь будет код для очистки сессий
	// Пока просто логируем действие
	log.Println("Clearing all sessions from Valkey")
	return nil
}
