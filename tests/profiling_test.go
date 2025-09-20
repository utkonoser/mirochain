package tests

import (
	"os"
	"testing"
	"time"

	"mirochain/internal/profiling"
)

// TestProfilerCreation тестирует создание профилировщика
func TestProfilerCreation(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6060",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	if profiler == nil {
		t.Fatal("Profiler should not be nil")
	}

	if !profiler.IsEnabled() {
		t.Error("Profiler should be enabled")
	}

	if profiler.GetProfileDir() != config.ProfileDir {
		t.Errorf("Expected profile dir %s, got %s", config.ProfileDir, profiler.GetProfileDir())
	}

	t.Logf("Profiler created successfully")
}

// TestProfilerStartStop тестирует запуск и остановку профилировщика
func TestProfilerStartStop(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6061", // Используем другой порт для тестов
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	// Запускаем профилировщик
	err := profiler.Start()
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}

	// Ждем немного, чтобы сервер запустился
	time.Sleep(100 * time.Millisecond)

	// Останавливаем профилировщик
	err = profiler.Stop()
	if err != nil {
		t.Fatalf("Failed to stop profiler: %v", err)
	}

	t.Logf("Profiler start/stop test completed successfully")
}

// TestProfilerDisabled тестирует отключенный профилировщик
func TestProfilerDisabled(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      false,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6062",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	if profiler.IsEnabled() {
		t.Error("Profiler should be disabled")
	}

	// Все операции должны возвращать nil для отключенного профилировщика
	err := profiler.Start()
	if err != nil {
		t.Errorf("Start should not return error for disabled profiler: %v", err)
	}

	err = profiler.Stop()
	if err != nil {
		t.Errorf("Stop should not return error for disabled profiler: %v", err)
	}

	err = profiler.WriteMemProfile("test.prof")
	if err != nil {
		t.Errorf("WriteMemProfile should not return error for disabled profiler: %v", err)
	}

	t.Logf("Disabled profiler test completed successfully")
}

// TestProfilerProfiles тестирует запись профилей
func TestProfilerProfiles(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6063",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	// Запускаем профилировщик
	err := profiler.Start()
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop()

	// Тестируем запись профилей
	err = profiler.WriteMemProfile("test_mem.prof")
	if err != nil {
		t.Errorf("Failed to write memory profile: %v", err)
	}

	err = profiler.WriteBlockProfile("test_block.prof")
	if err != nil {
		t.Errorf("Failed to write block profile: %v", err)
	}

	err = profiler.WriteMutexProfile("test_mutex.prof")
	if err != nil {
		t.Errorf("Failed to write mutex profile: %v", err)
	}

	err = profiler.WriteGoroutineProfile("test_goroutine.prof")
	if err != nil {
		t.Errorf("Failed to write goroutine profile: %v", err)
	}

	// Проверяем, что файлы созданы
	files := []string{
		"test_mem.prof",
		"test_block.prof",
		"test_mutex.prof",
		"test_goroutine.prof",
	}

	for _, file := range files {
		filePath := profiler.GetProfileDir() + "/" + file
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Profile file %s was not created", file)
		}
	}

	t.Logf("Profile writing test completed successfully")
}

// TestProfilerCPUProfile тестирует профилирование CPU
func TestProfilerCPUProfile(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6064",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	// Запускаем профилировщик
	err := profiler.Start()
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop()

	// Запускаем профилирование CPU
	err = profiler.StartCPUProfile("test_cpu.prof")
	if err != nil {
		t.Errorf("Failed to start CPU profile: %v", err)
	}

	// Выполняем некоторую работу
	time.Sleep(100 * time.Millisecond)

	// Останавливаем профилирование CPU
	err = profiler.StopCPUProfile()
	if err != nil {
		t.Errorf("Failed to stop CPU profile: %v", err)
	}

	// Проверяем, что файл создан
	filePath := profiler.GetProfileDir() + "/test_cpu.prof"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CPU profile file was not created")
	}

	t.Logf("CPU profiling test completed successfully")
}

// TestProfilerRuntimeStats тестирует получение статистики runtime
func TestProfilerRuntimeStats(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6065",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	// Получаем статистику runtime
	stats := profiler.GetRuntimeStats()

	// Проверяем, что статистика содержит ожидаемые поля
	expectedFields := []string{
		"num_goroutines",
		"num_cpu",
		"go_version",
		"mem_alloc_bytes",
		"mem_total_alloc",
		"mem_sys_bytes",
		"mem_heap_alloc",
		"mem_heap_sys",
		"mem_heap_idle",
		"mem_heap_inuse",
		"mem_heap_objects",
		"mem_stack_inuse",
		"mem_stack_sys",
		"mem_gc_cycles",
		"mem_gc_pause_total",
		"mem_gc_pause_ns",
	}

	for _, field := range expectedFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("Expected field %s not found in runtime stats", field)
		}
	}

	// Проверяем, что значения имеют правильный тип
	if _, ok := stats["num_goroutines"].(int); !ok {
		t.Error("num_goroutines should be int")
	}

	if _, ok := stats["num_cpu"].(int); !ok {
		t.Error("num_cpu should be int")
	}

	if _, ok := stats["go_version"].(string); !ok {
		t.Error("go_version should be string")
	}

	t.Logf("Runtime stats test completed successfully. Stats: %+v", stats)
}

// TestProfilerWriteAllProfiles тестирует запись всех профилей
func TestProfilerWriteAllProfiles(t *testing.T) {
	config := profiling.ProfilerConfig{
		Enabled:      true,
		ProfileDir:   t.TempDir(),
		HTTPPort:     ":6066",
		CPUProfile:   true,
		MemProfile:   true,
		BlockProfile: true,
		MutexProfile: true,
	}

	profiler := profiling.NewProfiler(config)

	// Запускаем профилировщик
	err := profiler.Start()
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer profiler.Stop()

	// Записываем все профили
	err = profiler.WriteAllProfiles("test_all")
	if err != nil {
		t.Errorf("Failed to write all profiles: %v", err)
	}

	// Проверяем, что файлы созданы
	files, err := os.ReadDir(profiler.GetProfileDir())
	if err != nil {
		t.Fatalf("Failed to read profile directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("No profile files were created")
	}

	t.Logf("Write all profiles test completed successfully. Created %d files", len(files))
}

// TestProfilerIntegration тестирует интеграцию профилировщика
func TestProfilerIntegration(t *testing.T) {
	t.Run("Creation", TestProfilerCreation)
	t.Run("StartStop", TestProfilerStartStop)
	t.Run("Disabled", TestProfilerDisabled)
	t.Run("Profiles", TestProfilerProfiles)
	t.Run("CPUProfile", TestProfilerCPUProfile)
	t.Run("RuntimeStats", TestProfilerRuntimeStats)
	t.Run("WriteAllProfiles", TestProfilerWriteAllProfiles)
}
