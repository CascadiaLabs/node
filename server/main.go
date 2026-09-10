package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	loadDotEnv(".env")

	token := os.Getenv("NODE_API_TOKEN")
	if token == "" {
		// fail-closed: без токена не поднимаем управляющий API
		log.Fatal("NODE_API_TOKEN не задан — отказ запуска (fail-closed)")
	}
	listenAddr := getenv("NODE_API_LISTEN", "0.0.0.0:6237")
	configPath := getenv("NODE_CONFIG_PATH", "/var/lib/node/config.json")

	mgr := NewManager(configPath)
	if err := mgr.Load(); err != nil {
		log.Fatalf("не удалось запустить sing-box: %v", err)
	}
	defer mgr.Close()

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("не удалось слушать %s: %v", listenAddr, err)
	}
	grpcSrv := NewGRPCServer(mgr, token)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("gRPC-сервер остановлен: %v", err)
		}
	}()
	fmt.Printf("🚀 Go-Нода запущена, gRPC API на %s\n", listenAddr)

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	<-osSignal

	fmt.Println("Остановка ноды...")
	grpcSrv.GracefulStop()
}

// getenv возвращает значение переменной окружения или fallback.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv загружает KEY=VALUE из файла, не перетирая уже заданные переменные окружения.
// Отсутствие файла — не ошибка.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
