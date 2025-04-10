package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ContainerIssue struct {
	Name   string
	ID     string
	Status string
	Error  string
}

type Monitor struct {
	dockerClient  *client.Client
	bot           *tgbotapi.BotAPI
	chatID        int64
	lastHeartbeat time.Time
	mutex         sync.Mutex
	isRunning     bool
}

func NewMonitor(dockerClient *client.Client, bot *tgbotapi.BotAPI, chatID int64) *Monitor {
	return &Monitor{
		dockerClient: dockerClient,
		bot:          bot,
		chatID:       chatID,
		isRunning:    true,
	}
}

func (m *Monitor) updateHeartbeat() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lastHeartbeat = time.Now()
}

func (m *Monitor) checkContainers() ([]ContainerIssue, error) {
	containers, err := m.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("error listando contenedores: %v", err)
	}

	var issues []ContainerIssue
	for _, container := range containers {
		if container.State != "running" && container.State != "created" {
			issue := ContainerIssue{
				Name:   container.Names[0],
				ID:     container.ID[:12],
				Status: container.State,
			}
			issues = append(issues, issue)
		}
	}

	return issues, nil
}

func (m *Monitor) sendTelegramMessage(message string) error {
	msg := tgbotapi.NewMessage(m.chatID, message)
	msg.ParseMode = "Markdown"
	_, err := m.bot.Send(msg)
	return err
}

func (m *Monitor) runHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			message := fmt.Sprintf("💚 Monitor activo - %s", time.Now().Format("2006-01-02 15:04:05"))
			if err := m.sendTelegramMessage(message); err != nil {
				log.Printf("Error enviando heartbeat: %v", err)
			}
			m.updateHeartbeat()
		}
	}
}

func (m *Monitor) runMonitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			issues, err := m.checkContainers()
			if err != nil {
				log.Printf("Error verificando contenedores: %v", err)
				m.sendTelegramMessage(fmt.Sprintf("⚠️ Error verificando contenedores: %v", err))
				continue
			}

			if len(issues) > 0 {
				message := fmt.Sprintf("🚨 *¡ALERTA!* Se encontraron %d contenedores con problemas:\n\n", len(issues))
				for _, issue := range issues {
					message += fmt.Sprintf("🔴 *Contenedor:* `%s`\n", issue.Name)
					message += fmt.Sprintf("📝 *ID:* `%s`\n", issue.ID)
					message += fmt.Sprintf("⚠️ *Estado:* `%s`\n", issue.Status)
					message += "------------------------\n"
				}
				if err := m.sendTelegramMessage(message); err != nil {
					log.Printf("Error enviando notificación: %v", err)
				}
			}
			m.updateHeartbeat()
		}
	}
}

func main() {
	// Configurar logging a stdout para Docker
	log.SetOutput(os.Stdout)

	// Obtener configuración desde variables de entorno
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Error: TELEGRAM_BOT_TOKEN no está configurado")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("Error: TELEGRAM_CHAT_ID no está configurado")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Error: TELEGRAM_CHAT_ID inválido: %v", err)
	}

	// Inicializar bot de Telegram
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error inicializando bot de Telegram: %v", err)
	}

	// Crear cliente Docker con configuración para contenedor
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHost("unix:///var/run/docker.sock"),
	)
	if err != nil {
		log.Fatalf("Error creando cliente Docker: %v", err)
	}
	defer cli.Close()

	// Crear contexto cancelable
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicializar monitor
	monitor := NewMonitor(cli, bot, chatID)

	// Manejar señales de sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar monitoreo en goroutines separadas
	go monitor.runMonitor(ctx)
	go monitor.runHeartbeat(ctx)

	// Enviar mensaje inicial
	monitor.sendTelegramMessage("🟢 Monitor de Docker iniciado. Vigilando contenedores...")

	// Esperar señal de terminación
	<-sigChan
	monitor.sendTelegramMessage("🔴 Monitor de Docker detenido")
	cancel()
}
