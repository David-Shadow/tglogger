package tests

import (
	"log"
	"testing"
	"time"

	"github.com/David-Shadow/tglogger"
)

func TestTgLogger(t *testing.T) {
	config := &tglogger.Config{
		Token:  "YOUR_BOT_TOKEN",
		ChatID: -100123456789,
		ExcludedLogPatterns: []string{"DEBUG", "FLOODWAIT"},
	}

	err := tglogger.InitializeTgLogger(config)
	if err != nil {
		panic(err)
	}

	log.Println("Testing if this thing actually runs XD")
    time.Sleep(500 * time.Millisecond)

	// These will be ignored due to ExcludedLogPatterns
	log.Println("DEBUG: This won't appear in tg")
	log.Println("FLOODWAIT: This is also ignored")
    time.Sleep(500 * time.Millisecond)

	// Testing with emoji
	log.Println("📈 Something happened here")
	log.Println("👤 New user started the bot: user123")
    time.Sleep(500 * time.Millisecond)

	// Simulate error
	log.Println("❌ ERROR: Damn! These errors!")
	log.Println("🔄 Errors everywhere 😭😭😭")
	log.Println("✅ Error resolved 😍😍😍😋😋")
    time.Sleep(500 * time.Millisecond)

	// File upload test
	log.Println("📊 Generating bulk logs for file upload test...")
	for i := range(400) {
		log.Printf("Bulk log entry #%d - %s", i, time.Now().Format("2006-01-02 15:04:05"))
		if i%20 == 0 {
			time.Sleep(100 * time.Millisecond) // Small delay
		}
	}

	log.Println("🏁 Demo completed! Wohooo🤩🤩🤩")

	// Wait for logs to be sent
	time.Sleep(5 * time.Second)
    log.Println("🏁 Did you receive the log file?🙂‍😋😏😏\npress Ctrl+C now😒😒😒")
}