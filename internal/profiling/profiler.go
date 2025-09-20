package profiling

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler представляет профилировщик узла
type Profiler struct {
	enabled    bool
	profileDir string
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
}

// ProfilerConfig представляет конфигурацию профилировщика
type ProfilerConfig struct {
	Enabled      bool
	ProfileDir   string
	HTTPPort     string
	CPUProfile   bool
	MemProfile   bool
	BlockProfile bool
	MutexProfile bool
}

// NewProfiler создает новый профилировщик
func NewProfiler(config ProfilerConfig) *Profiler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Profiler{
		enabled:    config.Enabled,
		profileDir: config.ProfileDir,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start запускает профилировщик
func (p *Profiler) Start() error {
	if !p.enabled {
		return nil
	}

	// Создаем директорию для профилей
	if err := os.MkdirAll(p.profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Запускаем HTTP сервер для pprof
	if p.httpServer == nil {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)

		p.httpServer = &http.Server{
			Addr:    ":6060",
			Handler: mux,
		}

		go func() {
			if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Failed to start pprof server", "error", err)
			}
		}()

		slog.Info("Profiler started", "port", ":6060", "profile_dir", p.profileDir)
	}

	return nil
}

// Stop останавливает профилировщик
func (p *Profiler) Stop() error {
	if !p.enabled {
		return nil
	}

	if p.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := p.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown pprof server: %w", err)
		}
	}

	if p.cancel != nil {
		p.cancel()
	}

	slog.Info("Profiler stopped")
	return nil
}

// StartCPUProfile запускает профилирование CPU
func (p *Profiler) StartCPUProfile(filename string) error {
	if !p.enabled {
		return nil
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", p.profileDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	slog.Info("CPU profiling started", "file", filename)
	return nil
}

// StopCPUProfile останавливает профилирование CPU
func (p *Profiler) StopCPUProfile() error {
	if !p.enabled {
		return nil
	}

	pprof.StopCPUProfile()
	slog.Info("CPU profiling stopped")
	return nil
}

// WriteMemProfile записывает профиль памяти
func (p *Profiler) WriteMemProfile(filename string) error {
	if !p.enabled {
		return nil
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", p.profileDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	slog.Info("Memory profile written", "file", filename)
	return nil
}

// WriteBlockProfile записывает профиль блокировок
func (p *Profiler) WriteBlockProfile(filename string) error {
	if !p.enabled {
		return nil
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", p.profileDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create block profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.Lookup("block").WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	slog.Info("Block profile written", "file", filename)
	return nil
}

// WriteMutexProfile записывает профиль мьютексов
func (p *Profiler) WriteMutexProfile(filename string) error {
	if !p.enabled {
		return nil
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", p.profileDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create mutex profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.Lookup("mutex").WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	slog.Info("Mutex profile written", "file", filename)
	return nil
}

// WriteGoroutineProfile записывает профиль горутин
func (p *Profiler) WriteGoroutineProfile(filename string) error {
	if !p.enabled {
		return nil
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", p.profileDir, filename))
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer file.Close()

	if err := pprof.Lookup("goroutine").WriteTo(file, 0); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	slog.Info("Goroutine profile written", "file", filename)
	return nil
}

// WriteAllProfiles записывает все профили
func (p *Profiler) WriteAllProfiles(prefix string) error {
	if !p.enabled {
		return nil
	}

	timestamp := time.Now().Format("20060102_150405")

	// Профиль памяти
	if err := p.WriteMemProfile(fmt.Sprintf("%s_mem_%s.prof", prefix, timestamp)); err != nil {
		slog.Error("Failed to write memory profile", "error", err)
	}

	// Профиль блокировок
	if err := p.WriteBlockProfile(fmt.Sprintf("%s_block_%s.prof", prefix, timestamp)); err != nil {
		slog.Error("Failed to write block profile", "error", err)
	}

	// Профиль мьютексов
	if err := p.WriteMutexProfile(fmt.Sprintf("%s_mutex_%s.prof", prefix, timestamp)); err != nil {
		slog.Error("Failed to write mutex profile", "error", err)
	}

	// Профиль горутин
	if err := p.WriteGoroutineProfile(fmt.Sprintf("%s_goroutine_%s.prof", prefix, timestamp)); err != nil {
		slog.Error("Failed to write goroutine profile", "error", err)
	}

	slog.Info("All profiles written", "prefix", prefix, "timestamp", timestamp)
	return nil
}

// GetRuntimeStats возвращает статистику runtime
func (p *Profiler) GetRuntimeStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"num_goroutines":     runtime.NumGoroutine(),
		"num_cpu":            runtime.NumCPU(),
		"go_version":         runtime.Version(),
		"mem_alloc_bytes":    m.Alloc,
		"mem_total_alloc":    m.TotalAlloc,
		"mem_sys_bytes":      m.Sys,
		"mem_heap_alloc":     m.HeapAlloc,
		"mem_heap_sys":       m.HeapSys,
		"mem_heap_idle":      m.HeapIdle,
		"mem_heap_inuse":     m.HeapInuse,
		"mem_heap_objects":   m.HeapObjects,
		"mem_stack_inuse":    m.StackInuse,
		"mem_stack_sys":      m.StackSys,
		"mem_gc_cycles":      m.NumGC,
		"mem_gc_pause_total": time.Duration(m.PauseTotalNs),
		"mem_gc_pause_ns":    m.PauseNs,
	}
}

// LogRuntimeStats логирует статистику runtime
func (p *Profiler) LogRuntimeStats(component string) {
	stats := p.GetRuntimeStats()

	slog.Info("Runtime stats",
		"component", component,
		"goroutines", stats["num_goroutines"],
		"cpu_count", stats["num_cpu"],
		"go_version", stats["go_version"],
		"mem_alloc_mb", stats["mem_alloc_bytes"].(uint64)/1024/1024,
		"mem_heap_alloc_mb", stats["mem_heap_alloc"].(uint64)/1024/1024,
		"mem_heap_objects", stats["mem_heap_objects"],
		"gc_cycles", stats["mem_gc_cycles"],
	)
}

// StartPeriodicProfiling запускает периодическое профилирование
func (p *Profiler) StartPeriodicProfiling(interval time.Duration, prefix string) {
	if !p.enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := p.WriteAllProfiles(prefix); err != nil {
					slog.Error("Failed to write periodic profiles", "error", err)
				}
			case <-p.ctx.Done():
				return
			}
		}
	}()

	slog.Info("Periodic profiling started", "interval", interval, "prefix", prefix)
}

// StartPeriodicStats запускает периодическое логирование статистики
func (p *Profiler) StartPeriodicStats(interval time.Duration, component string) {
	if !p.enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.LogRuntimeStats(component)
			case <-p.ctx.Done():
				return
			}
		}
	}()

	slog.Info("Periodic stats logging started", "interval", interval, "component", component)
}

// IsEnabled возвращает, включен ли профилировщик
func (p *Profiler) IsEnabled() bool {
	return p.enabled
}

// GetProfileDir возвращает директорию профилей
func (p *Profiler) GetProfileDir() string {
	return p.profileDir
}

// GetHTTPPort возвращает порт HTTP сервера
func (p *Profiler) GetHTTPPort() string {
	if p.httpServer != nil {
		return p.httpServer.Addr
	}
	return ""
}
