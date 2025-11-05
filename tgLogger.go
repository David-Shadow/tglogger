package tglogger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Token               string
	ChatID              int64
	ForumTopicID        int
	Title               string
	ExcludedLogPatterns []string
	UpdateInterval      time.Duration
	MinimumLines        int
	PendingLogsSize     int
	MaxMessageSize      int
}

type TelegramLogger struct {
	config  *Config
	client  *http.Client
	baseURL string

	mu            sync.Mutex
	logBuffer     strings.Builder
	currentMsg    string
	messageID     int
	lines         int
	lastLogUpdate time.Time
	floodWait     time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	logFile *os.File
}

type TelegramResponse struct {
	OK          bool           `json:"ok"`
	Result      map[string]any `json:"result,omitempty"`
	ErrorCode   int            `json:"error_code,omitempty"`
	Description string         `json:"description,omitempty"`
	Parameters  struct {
		RetryAfter int `json:"retry_after,omitempty"`
	} `json:"parameters,omitempty"`
}

func InitializeTgLogger(config *Config) error {
	if config.ChatID == 0 {
		return fmt.Errorf("please provide ChatID")
	}

	if config.Token == "" {
		return fmt.Errorf("please provide Token")
	}

	if config.UpdateInterval <= 0 {
		config.UpdateInterval = 3 * time.Second
	}

	if config.MinimumLines <= 0 {
		config.MinimumLines = 1
	}

	if config.PendingLogsSize <= 0 {
		config.PendingLogsSize = 20000
	}

	if config.MaxMessageSize <= 0 {
		config.MaxMessageSize = 4096
	}

	if config.Title == "" {
		config.Title = "TGLogger-Go"
	}

	if config.ForumTopicID < 0 {
		config.ForumTopicID = 0
	}

	logFileName := "bot.log"
	os.Truncate(logFileName, 0)
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Print("Failed to open log file: ", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := &TelegramLogger{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", config.Token),
		ctx:     ctx,
		cancel:  cancel,
		logFile: file,
	}

	if username, err := logger.validateBotToken(); err != nil {
		return err
	} else {
		fmt.Printf("Using @%s for tglogger!\n", username)
	}
	log.SetOutput(logger)

	return nil
}

func (logger *TelegramLogger) Write(p []byte) (n int, err error) {
	msg := string(p)

	n, err = logger.logFile.Write(p)
	if err != nil {
		return n, err
	}

	n, err = os.Stdout.Write(p)
	if err != nil {
		return n, err
	}

	for _, pattern := range logger.config.ExcludedLogPatterns {
		if strings.Contains(msg, pattern) {
			return len(p), nil
		}
	}

	logger.mu.Lock()
	logger.logBuffer.WriteString(msg)
	if !strings.HasSuffix(msg, "\n") {
		logger.logBuffer.WriteString("\n")
	}
	logger.lines++
	shouldUpdate := time.Since(logger.lastLogUpdate) >= max(logger.config.UpdateInterval, logger.floodWait) &&
		logger.lines >= logger.config.MinimumLines &&
		logger.logBuffer.Len() > 0
		
	if shouldUpdate {
		if err := logger.sendLogs(); err != nil {
			fmt.Printf("[TGLogger] Error handling logs: %v", err)
		}
	}
	logger.mu.Unlock()

	return len(p), nil
}

func (logger *TelegramLogger) sendLogs() error {
	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.floodWait > 0 {
		logger.floodWait = 0
	}

	if logger.logBuffer.Len() == 0 {
		return nil
	}

	logBuffer := logger.logBuffer.String()
	if len(logBuffer) == 0 {
		return nil
	}

	if len(logBuffer) > logger.config.PendingLogsSize {
		lastNewline := strings.LastIndex(logBuffer, "\n")
		if lastNewline == -1 {
			lastNewline = len(logBuffer)
		}

		toSend := logBuffer[:lastNewline]
		logger.logBuffer.Reset()
		logger.logBuffer.WriteString(logBuffer[lastNewline:])
		logger.currentMsg = ""
		logger.messageID = 0
		logger.lines = 0
		logger.lastLogUpdate = time.Now()

		return logger.sendAsFile(toSend)
	}

	maxLen := min(4000, len(logBuffer))
	toProcess := logBuffer[:maxLen]
	lastNewline := strings.LastIndex(toProcess, "\n")
	if lastNewline == -1 {
		lastNewline = len(toProcess)
	}

	msg := toProcess[:lastNewline]
	if len(msg) == 0 {
		return nil
	}

	remaining := logBuffer[len(msg):]
	logger.logBuffer.Reset()
	logger.logBuffer.WriteString(remaining)
	logger.lines = 0
	logger.lastLogUpdate = time.Now()

	if logger.messageID == 0 {
		if err := logger.initialize(); err != nil {
			return err
		}
	}

	computedMessage := logger.currentMsg + msg

	if len(computedMessage) > 4000 {
		lastNewlineInComputed := strings.LastIndex(computedMessage[:4000], "\n")
		if lastNewlineInComputed == -1 {
			lastNewlineInComputed = 4000
		}

		toEdit := computedMessage[:lastNewlineInComputed]
		toNew := computedMessage[lastNewlineInComputed:]

		if toEdit != logger.currentMsg {
			if err := logger.editTgMessage(toEdit); err != nil {
				return err
			}
		}

		logger.currentMsg = toNew
		return logger.sendTgMessage(toNew)
	}

	if err := logger.editTgMessage(computedMessage); err != nil {
		return err
	}
	logger.currentMsg = computedMessage

	return nil
}

func (logger *TelegramLogger) initialize() error {
	payload := map[string]any{
		"chat_id":                  logger.config.ChatID,
		"text":                     "```\nInitializing TGLogger-Go\n```",
		"parse_mode":               "Markdown",
		"disable_web_page_preview": true,
	}

	if logger.config.ForumTopicID > 0 {
		payload["message_thread_id"] = logger.config.ForumTopicID
	}

	resp, err := logger.makeTelegramRequest("sendMessage", payload)
	if err != nil {
		return err
	}

	if resp.OK {
		if msgID, ok := resp.Result["message_id"].(float64); ok {
			logger.messageID = int(msgID)
		}
	} else {
		return logger.handleError(resp)
	}

	return nil
}

func (logger *TelegramLogger) sendTgMessage(message string) error {
	text := fmt.Sprintf("```\n%s\n\n%s\n```", logger.config.Title, message)

	payload := map[string]interface{}{
		"chat_id":                  logger.config.ChatID,
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": true,
	}

	if logger.config.ForumTopicID > 0 {
		payload["message_thread_id"] = logger.config.ForumTopicID
	}

	resp, err := logger.makeTelegramRequest("sendMessage", payload)
	if err != nil {
		return err
	}

	if resp.OK {
		if msgID, ok := resp.Result["message_id"].(float64); ok {
			logger.messageID = int(msgID)
		}
	} else {
		return logger.handleError(resp)
	}

	return nil
}

func (logger *TelegramLogger) editTgMessage(message string) error {
	text := fmt.Sprintf("```\n%s\n\n%s\n```", logger.config.Title, message)

	payload := map[string]interface{}{
		"chat_id":                  logger.config.ChatID,
		"message_id":               logger.messageID,
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": true,
	}

	resp, err := logger.makeTelegramRequest("editMessageText", payload)
	if err != nil {
		return err
	}

	if !resp.OK {
		return logger.handleError(resp)
	}

	return nil
}

func (logger *TelegramLogger) sendAsFile(logs string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("document", logger.logFile.Name())
	if err != nil {
		return err
	}

	if _, err := fileWriter.Write([]byte(logs)); err != nil {
		return err
	}

	if err := writer.WriteField("chat_id", fmt.Sprintf("%d", logger.config.ChatID)); err != nil {
		return err
	}

	if err := writer.WriteField("caption", "Too many logs for text logBuffer! This file contains the logs."); err != nil {
		return err
	}

	if logger.config.ForumTopicID > 0 {
		if err := writer.WriteField("message_thread_id", fmt.Sprintf("%d", logger.config.ForumTopicID)); err != nil {
			return err
		}
	}

	writer.Close()

	req, err := http.NewRequestWithContext(logger.ctx, "POST", logger.baseURL+"/sendDocument", &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := logger.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return err
	}

	if telegramResp.OK {
		fmt.Println("[TGLogger] Sent logs as file due to size limit")
	} else {
		return logger.handleError(&telegramResp)
	}

	return nil
}

func (logger *TelegramLogger) makeTelegramRequest(method string, payload map[string]interface{}) (*TelegramResponse, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(logger.ctx, "POST", logger.baseURL+"/"+method, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := logger.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		return nil, err
	}

	return &telegramResp, nil
}

func (logger *TelegramLogger) handleError(resp *TelegramResponse) error {
	if resp.ErrorCode == 401 && resp.Description == "Unauthorized" {
		return fmt.Errorf("unauthorized: invalid bot token")
	}

	if resp.Parameters.RetryAfter > 0 {
		logger.floodWait = time.Duration(resp.Parameters.RetryAfter) * time.Second
		fmt.Printf("[TGLogger] Got FloodWait of %d seconds, sleeping...", resp.Parameters.RetryAfter)
		return nil
	}

	return fmt.Errorf("telegram API error: %d - %s", resp.ErrorCode, resp.Description)
}

func (logger *TelegramLogger) validateBotToken() (string, error) {
	resp, err := logger.makeTelegramRequest("getMe", map[string]interface{}{})
	if err != nil {
		return "", err
	}

	if !resp.OK {
		return "", fmt.Errorf("bot verification failed: %s", resp.Description)
	}

	if username, ok := resp.Result["username"].(string); ok {
		return username, nil
	}

	return "", fmt.Errorf("unable to get bot username")
}
