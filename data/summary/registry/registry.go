package registry

import (
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/types"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type SummarizerRegistry struct {
	summarizers map[string]any
}

func New(summaryRepository *storeStructuredMongo.Repository, dataRepository dataStore.DataRepository) *SummarizerRegistry {
	registry := &SummarizerRegistry{summarizers: make(map[string]any)}
	addSummarizer(registry, NewCGMSummarizer(summaryRepository, dataRepository))
	addSummarizer(registry, NewBGMSummarizer(summaryRepository, dataRepository))
	return registry
}

func addSummarizer[T types.Stats](reg *SummarizerRegistry, summarizer Summarizer[T]) {
	typ := types.GetTypeString[T]()
	reg.summarizers[typ] = summarizer
}

func GetSummarizer[T types.Stats](reg *SummarizerRegistry) Summarizer[T] {
	typ := types.GetTypeString[T]()
	summarizer := reg.summarizers[typ]
	return summarizer.(Summarizer[T])
}
