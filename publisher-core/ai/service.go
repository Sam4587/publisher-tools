package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"publisher-core/ai/provider"
	"github.com/sirupsen/logrus"
)

type Service struct {
	mu        sync.RWMutex
	providers map[provider.ProviderType]provider.Provider
	config    *Config
	primary   provider.ProviderType
}

type Config struct {
	Primary   provider.ProviderType     `json:"primary"`
	Providers map[string]ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"`
	Model    string `json:"default_model,omitempty"`
	Enabled  bool   `json:"enabled"`
	Priority int    `json:"priority"`
}

func NewService(configPath string) (*Service, error) {
	s := &Service{
		providers: make(map[provider.ProviderType]provider.Provider),
		primary:   provider.ProviderOpenRouter,
	}

	if configPath != "" {
		if err := s.loadConfig(configPath); err != nil {
			logrus.Warnf("load AI config failed: %v, using defaults", err)
		}
	}

	return s, nil
}

func NewServiceWithDefaults() *Service {
	s := &Service{
		providers: make(map[provider.ProviderType]provider.Provider),
		primary:   provider.ProviderOpenRouter,
		config:    &Config{Primary: provider.ProviderOpenRouter, Providers: make(map[string]ProviderConfig)},
	}
	return s
}

func (s *Service) loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	s.config = &cfg
	s.primary = cfg.Primary

	for name, pc := range cfg.Providers {
		if !pc.Enabled || pc.APIKey == "" {
			continue
		}

		pt := provider.ProviderType(name)
		switch pt {
		case provider.ProviderOpenRouter:
			s.RegisterProvider(provider.NewOpenRouterProvider(pc.APIKey))
		case provider.ProviderGoogle:
			s.RegisterProvider(provider.NewGoogleProvider(pc.APIKey))
		case provider.ProviderGroq:
			s.RegisterProvider(provider.NewGroqProvider(pc.APIKey))
		case provider.ProviderDeepSeek:
			p := provider.NewDeepSeekProviderWithBaseURL(pc.APIKey, pc.BaseURL)
			s.RegisterProvider(p)
		}
	}

	return nil
}

func (s *Service) RegisterProvider(p provider.Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[p.Name()] = p
	logrus.Infof("AI provider registered: %s", p.Name())
}

func (s *Service) SetPrimary(pt provider.ProviderType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.providers[pt]; !ok {
		return fmt.Errorf("provider not registered: %s", pt)
	}

	s.primary = pt
	logrus.Infof("Primary AI provider set to: %s", pt)
	return nil
}

func (s *Service) GetPrimary() provider.Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if p, ok := s.providers[s.primary]; ok {
		return p
	}

	for _, p := range s.providers {
		return p
	}

	return nil
}

func (s *Service) GetProvider(pt provider.ProviderType) (provider.Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.providers[pt]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", pt)
	}
	return p, nil
}

func (s *Service) ListProviders() []provider.ProviderType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]provider.ProviderType, 0, len(s.providers))
	for pt := range s.providers {
		result = append(result, pt)
	}
	return result
}

func (s *Service) ListModels() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]string)
	for pt, p := range s.providers {
		result[string(pt)] = p.Models()
	}
	return result
}

func (s *Service) Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	p := s.GetPrimary()
	if p == nil {
		return nil, fmt.Errorf("no AI provider available")
	}

	if opts.Model == "" {
		opts.Model = p.DefaultModel()
	}

	return p.Generate(ctx, opts)
}

func (s *Service) GenerateStream(ctx context.Context, opts *provider.GenerateOptions) (<-chan string, error) {
	p := s.GetPrimary()
	if p == nil {
		return nil, fmt.Errorf("no AI provider available")
	}

	if opts.Model == "" {
		opts.Model = p.DefaultModel()
	}

	return p.GenerateStream(ctx, opts)
}

func (s *Service) GenerateWithProvider(ctx context.Context, pt provider.ProviderType, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	p, err := s.GetProvider(pt)
	if err != nil {
		return nil, err
	}

	if opts.Model == "" {
		opts.Model = p.DefaultModel()
	}

	return p.Generate(ctx, opts)
}

func SaveConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func DefaultConfigPath() string {
	return filepath.Join(".", "config", "ai.json")
}
