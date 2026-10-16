package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram struct {
		BotToken string `yaml:"bot_token"`
		ChatID   string `yaml:"chat_id"`
	} `yaml:"telegram"`
	CheckInterval int `yaml:"check_interval_seconds"`
	Targets       []struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"targets"`
}

type Monitor struct {
	config Config
	downState map[string]bool
	mu        sync.RWMutex
}

var (
	upMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "uptime_target_up", Help: "1 if target is up, 0 if down"},
		[]string{"name", "url"},
	)
	durationMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "uptime_check_duration_seconds", Help: "Time taken to check target"},
		[]string{"name", "url"},
	)
)

func init() {
	prometheus.MustRegister(upMetric)
	prometheus.MustRegister(durationMetric)
}

func main() {
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Не удалось прочитать config.yaml: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		log.Fatalf("Ошибка парсинга YAML: %v", err)
	}

	monitor := &Monitor{
		config:    cfg,
		downState: make(map[string]bool),
	}

	log.Printf("Запущен Uptime Monitor. Интервал: %d сек. Целей: %d", cfg.CheckInterval, len(cfg.Targets))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📊 Метрики доступны на http://localhost:9116/metrics")
		log.Fatal(http.ListenAndServe(":9116", nil))
	}()

	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	defer ticker.Stop()
	monitor.checkAll()

	for range ticker.C {
		monitor.checkAll()
	}
}

func (m *Monitor) checkAll() {
	var wg sync.WaitGroup

	for _, target := range m.config.Targets {
		wg.Add(1)
		go func(t struct{ Name, URL string }) {
			defer wg.Done()
			m.checkTarget(t)
		}(target)
	}

	wg.Wait()
}

func (m *Monitor) checkTarget(target struct{ Name, URL string }) {
	start := time.Now()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(target.URL)
	duration := time.Since(start).Seconds()

	durationMetric.WithLabelValues(target.Name, target.URL).Set(duration)

	var isDown bool
	if err != nil || resp.StatusCode >= 500 {
		isDown = true
		log.Printf("DOWN: %s (%s) - %v", target.Name, target.URL, err)
	} else {
		isDown = false
		log.Printf("UP:   %s (%s) - %d (%.2fs)", target.Name, target.URL, resp.StatusCode, duration)
		resp.Body.Close()
	}

	if isDown {
		upMetric.WithLabelValues(target.Name, target.URL).Set(0)
	} else {
		upMetric.WithLabelValues(target.Name, target.URL).Set(1)
	}

	m.mu.Lock()
prevState := m.downState[target.URL]
	if prevState != isDown {
		m.downState[target.URL] = isDown
		if isDown {
			m.sendTelegramAlert(target.Name, target.URL, "DOWN", err)
		} else {
			m.sendTelegramAlert(target.Name, target.URL, "UP", nil)
		}
	}
	m.mu.Unlock()
}

func (m *Monitor) sendTelegramAlert(name, url, status string, err error) {
	if m.config.Telegram.BotToken == "" || m.config.Telegram.ChatID == "" {
		log.Println("Telegram не настроен, пропускаем отправку алерта")
		return
	}

	msg := fmt.Sprintf("*%s* %s\n URL: `%s`\n Время: %s", 
		name, status, url, time.Now().Format("15:04:05"))
	
	if err != nil {
		msg += fmt.Sprintf("\n Ошибка: `%v`", err)
	}

	payload := map[string]string{
		"chat_id":                  m.config.Telegram.ChatID,
		"text":                     msg,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": "true",
	}

	jsonData, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.config.Telegram.BotToken)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Ошибка отправки в Telegram: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		log.Printf("Алерт отправлен в Telegram: %s %s", name, status)
	} else {
		log.Printf("Telegram вернул статус %d", resp.StatusCode)
	}
}