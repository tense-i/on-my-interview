package crawler

import "fmt"

type Registry struct {
	crawlers map[string]PlatformCrawler
}

func NewRegistry(items ...PlatformCrawler) *Registry {
	registry := &Registry{crawlers: make(map[string]PlatformCrawler, len(items))}
	for _, item := range items {
		registry.crawlers[item.Platform()] = item
	}
	return registry
}

func (r *Registry) Get(platform string) (PlatformCrawler, error) {
	crawler, ok := r.crawlers[platform]
	if !ok {
		return nil, fmt.Errorf("crawler for platform %q not registered", platform)
	}
	return crawler, nil
}
