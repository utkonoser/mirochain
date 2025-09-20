package gateway

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// VersionManager управляет версионированием API
type VersionManager struct {
	versions map[string]*APIVersion
}

// APIVersion представляет версию API
type APIVersion struct {
	Version     string            `json:"version"`
	Status      VersionStatus     `json:"status"`
	ReleaseDate time.Time         `json:"release_date"`
	Deprecated  time.Time         `json:"deprecated,omitempty"`
	Sunset      time.Time         `json:"sunset,omitempty"`
	Features    []string          `json:"features"`
	Changes     []string          `json:"changes"`
	Endpoints   map[string]string `json:"endpoints"`
	Documentation string          `json:"documentation"`
}

// VersionStatus представляет статус версии
type VersionStatus string

const (
	StatusDevelopment VersionStatus = "development"
	StatusStable      VersionStatus = "stable"
	StatusDeprecated  VersionStatus = "deprecated"
	StatusSunset      VersionStatus = "sunset"
)

// NewVersionManager создает новый VersionManager
func NewVersionManager() *VersionManager {
	vm := &VersionManager{
		versions: make(map[string]*APIVersion),
	}
	
	// Инициализируем базовые версии
	vm.initializeVersions()
	
	return vm
}

// initializeVersions инициализирует базовые версии API
func (vm *VersionManager) initializeVersions() {
	// Версия 1.0 - Базовая функциональность
	vm.versions["v1"] = &APIVersion{
		Version:     "v1",
		Status:      StatusStable,
		ReleaseDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Features: []string{
			"Basic blockchain queries",
			"Transaction management",
			"Wallet operations",
			"Smart contract deployment",
			"REST API endpoints",
			"Basic GraphQL support",
		},
		Changes: []string{
			"Initial release",
			"Core blockchain functionality",
			"Basic smart contracts",
		},
		Endpoints: map[string]string{
			"blockchain": "/api/v1/blockchain",
			"blocks":     "/api/v1/blocks",
			"transactions": "/api/v1/transactions",
			"wallets":    "/api/v1/wallets",
			"contracts":  "/api/v1/contracts",
			"graphql":    "/graphql",
		},
		Documentation: "https://docs.mirochain.com/v1",
	}
	
	// Версия 2.0 - Расширенная функциональность
	vm.versions["v2"] = &APIVersion{
		Version:     "v2",
		Status:      StatusStable,
		ReleaseDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		Features: []string{
			"Enhanced blockchain queries",
			"Advanced smart contract support",
			"Token management (ERC-20)",
			"NFT support (ERC-721)",
			"Sidechain integration",
			"State channels",
			"WebSocket support",
			"Webhook system",
			"Advanced GraphQL",
			"Rate limiting",
			"Input validation",
		},
		Changes: []string{
			"Added token system",
			"Added NFT support",
			"Added sidechain functionality",
			"Added state channels",
			"Enhanced security features",
			"Improved performance",
			"Better error handling",
		},
		Endpoints: map[string]string{
			"blockchain": "/api/v2/blockchain",
			"blocks":     "/api/v2/blocks",
			"transactions": "/api/v2/transactions",
			"wallets":    "/api/v2/wallets",
			"contracts":  "/api/v2/contracts",
			"tokens":     "/api/v2/tokens",
			"nfts":       "/api/v2/nfts",
			"sidechains": "/api/v2/sidechains",
			"channels":   "/api/v2/channels",
			"webhooks":   "/api/v2/webhooks",
			"graphql":    "/graphql",
		},
		Documentation: "https://docs.mirochain.com/v2",
	}
	
	// Версия 3.0 - Будущая версия (в разработке)
	vm.versions["v3"] = &APIVersion{
		Version:     "v3",
		Status:      StatusDevelopment,
		ReleaseDate: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Features: []string{
			"Quantum-resistant cryptography",
			"Advanced consensus algorithms",
			"Cross-chain interoperability",
			"Privacy features",
			"Scalability improvements",
			"Advanced analytics",
			"Machine learning integration",
			"Real-time monitoring",
		},
		Changes: []string{
			"Quantum-resistant algorithms",
			"Enhanced privacy",
			"Better scalability",
			"Cross-chain support",
		},
		Endpoints: map[string]string{
			"blockchain": "/api/v3/blockchain",
			"quantum":    "/api/v3/quantum",
			"privacy":    "/api/v3/privacy",
			"crosschain": "/api/v3/crosschain",
			"analytics":  "/api/v3/analytics",
			"graphql":    "/graphql",
		},
		Documentation: "https://docs.mirochain.com/v3",
	}
	
	// Настраиваем deprecated и sunset даты
	vm.versions["v1"].Deprecated = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	vm.versions["v1"].Sunset = time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
}

// GetVersion возвращает информацию о версии
func (vm *VersionManager) GetVersion(version string) (*APIVersion, error) {
	if v, exists := vm.versions[version]; exists {
		return v, nil
	}
	return nil, fmt.Errorf("version %s not found", version)
}

// GetLatestVersion возвращает последнюю стабильную версию
func (vm *VersionManager) GetLatestVersion() *APIVersion {
	var latest *APIVersion
	var latestDate time.Time
	
	for _, version := range vm.versions {
		if version.Status == StatusStable && version.ReleaseDate.After(latestDate) {
			latest = version
			latestDate = version.ReleaseDate
		}
	}
	
	return latest
}

// ListVersions возвращает список всех версий
func (vm *VersionManager) ListVersions() []*APIVersion {
	versions := make([]*APIVersion, 0, len(vm.versions))
	for _, version := range vm.versions {
		versions = append(versions, version)
	}
	
	// Сортируем по дате релиза
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].ReleaseDate.After(versions[j].ReleaseDate)
	})
	
	return versions
}

// IsDeprecated проверяет, устарела ли версия
func (vm *VersionManager) IsDeprecated(version string) bool {
	v, err := vm.GetVersion(version)
	if err != nil {
		return false
	}
	
	return !v.Deprecated.IsZero() && time.Now().After(v.Deprecated)
}

// IsSunset проверяет, завершена ли поддержка версии
func (vm *VersionManager) IsSunset(version string) bool {
	v, err := vm.GetVersion(version)
	if err != nil {
		return false
	}
	
	return !v.Sunset.IsZero() && time.Now().After(v.Sunset)
}

// GetSupportedVersions возвращает поддерживаемые версии
func (vm *VersionManager) GetSupportedVersions() []*APIVersion {
	var supported []*APIVersion
	
	for _, version := range vm.versions {
		if version.Status == StatusStable && !vm.IsSunset(version.Version) {
			supported = append(supported, version)
		}
	}
	
	return supported
}

// GetVersionHeaders возвращает заголовки для версии
func (vm *VersionManager) GetVersionHeaders(version string) map[string]string {
	v, err := vm.GetVersion(version)
	if err != nil {
		return map[string]string{}
	}
	
	headers := map[string]string{
		"X-API-Version": version,
		"X-API-Status":  string(v.Status),
		"X-API-Release-Date": v.ReleaseDate.Format(time.RFC3339),
	}
	
	if !v.Deprecated.IsZero() {
		headers["X-API-Deprecated"] = "true"
		headers["X-API-Deprecated-Date"] = v.Deprecated.Format(time.RFC3339)
	}
	
	if !v.Sunset.IsZero() {
		headers["X-API-Sunset"] = "true"
		headers["X-API-Sunset-Date"] = v.Sunset.Format(time.RFC3339)
	}
	
	return headers
}

// GetMigrationPath возвращает путь миграции между версиями
func (vm *VersionManager) GetMigrationPath(from, to string) ([]string, error) {
	fromVersion, err := vm.GetVersion(from)
	if err != nil {
		return nil, fmt.Errorf("source version %s not found", from)
	}
	
	toVersion, err := vm.GetVersion(to)
	if err != nil {
		return nil, fmt.Errorf("target version %s not found", to)
	}
	
	// Простая логика миграции
	var steps []string
	
	if fromVersion.Version == "v1" && toVersion.Version == "v2" {
		steps = []string{
			"Update API endpoints to v2",
			"Add token management functionality",
			"Implement NFT support",
			"Add sidechain integration",
			"Update authentication headers",
			"Test new features",
		}
	} else if fromVersion.Version == "v2" && toVersion.Version == "v3" {
		steps = []string{
			"Update to quantum-resistant cryptography",
			"Implement privacy features",
			"Add cross-chain support",
			"Update security protocols",
			"Test new algorithms",
		}
	} else {
		steps = []string{
			"Review version differences",
			"Update API calls",
			"Test compatibility",
			"Update documentation",
		}
	}
	
	return steps, nil
}

// GetVersionComparison возвращает сравнение версий
func (vm *VersionManager) GetVersionComparison(version1, version2 string) (map[string]interface{}, error) {
	v1, err := vm.GetVersion(version1)
	if err != nil {
		return nil, fmt.Errorf("version %s not found", version1)
	}
	
	v2, err := vm.GetVersion(version2)
	if err != nil {
		return nil, fmt.Errorf("version %s not found", version2)
	}
	
	comparison := map[string]interface{}{
		"version1": v1,
		"version2": v2,
		"differences": map[string]interface{}{
			"features_added": vm.getAddedFeatures(v1, v2),
			"features_removed": vm.getRemovedFeatures(v1, v2),
			"endpoints_changed": vm.getChangedEndpoints(v1, v2),
		},
	}
	
	return comparison, nil
}

// getAddedFeatures возвращает добавленные функции
func (vm *VersionManager) getAddedFeatures(v1, v2 *APIVersion) []string {
	var added []string
	
	for _, feature := range v2.Features {
		found := false
		for _, v1Feature := range v1.Features {
			if feature == v1Feature {
				found = true
				break
			}
		}
		if !found {
			added = append(added, feature)
		}
	}
	
	return added
}

// getRemovedFeatures возвращает удаленные функции
func (vm *VersionManager) getRemovedFeatures(v1, v2 *APIVersion) []string {
	var removed []string
	
	for _, feature := range v1.Features {
		found := false
		for _, v2Feature := range v2.Features {
			if feature == v2Feature {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, feature)
		}
	}
	
	return removed
}

// getChangedEndpoints возвращает измененные endpoints
func (vm *VersionManager) getChangedEndpoints(v1, v2 *APIVersion) map[string]string {
	changed := make(map[string]string)
	
	for endpoint, v1URL := range v1.Endpoints {
		if v2URL, exists := v2.Endpoints[endpoint]; exists {
			if v1URL != v2URL {
				changed[endpoint] = fmt.Sprintf("%s -> %s", v1URL, v2URL)
			}
		}
	}
	
	return changed
}

// ValidateVersion проверяет валидность версии
func (vm *VersionManager) ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version is required")
	}
	
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version must start with 'v'")
	}
	
	_, err := vm.GetVersion(version)
	return err
}

// GetVersionInfo возвращает краткую информацию о версии
func (vm *VersionManager) GetVersionInfo(version string) map[string]interface{} {
	v, err := vm.GetVersion(version)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	return map[string]interface{}{
		"version": v.Version,
		"status": v.Status,
		"deprecated": vm.IsDeprecated(version),
		"sunset": vm.IsSunset(version),
		"supported": !vm.IsSunset(version),
		"features_count": len(v.Features),
		"endpoints_count": len(v.Endpoints),
	}
}
