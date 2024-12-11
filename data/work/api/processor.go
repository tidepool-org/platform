package work

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataStore "github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
)

const (
	Type              = "org.tidepool.work.api.ingest"
	ProcessingTimeout = 300
)

func NewIngestCreate(dataSet *data.DataSet, raw *dataRaw.Raw) *work.Create {
	create := work.NewCreate()
	create.Type = Type
	create.GroupID = dataSet.ID
	create.DeduplicationID = pointer.FromString(raw.DigestMD5)
	create.SerialID = dataSet.ID
	create.ProcessingTimeout = ProcessingTimeout
	create.Metadata = metadata.NewMetadata()
	create.Metadata.Set("dataSetId", dataSet.ID)
	create.Metadata.Set("rawId", raw.ID)
	return create
}

type IngestProcessorProvider interface {
	DataRawClient() dataRaw.Client
	DataRepository() dataStore.DataRepository
	DataDeduplicatorFactory() deduplicator.Factory
	SummarizerRegistry() *summary.SummarizerRegistry
}

func NewIngestProcessor(provider IngestProcessorProvider) (*IngestProcessor, error) {
	if provider == nil {
		return nil, errors.New("provider is missing")
	}
	return &IngestProcessor{
		IngestProcessorProvider: provider,
	}, nil
}

type IngestProcessor struct {
	IngestProcessorProvider
}

func (i *IngestProcessor) Type() string {
	return Type
}

func (i *IngestProcessor) Quantity() int {
	return 8 // TODO: Configuration?
}

func (i *IngestProcessor) Frequency() time.Duration {
	return 5 * time.Second // TODO: Configuration?
}

func (i *IngestProcessor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) *work.PendingUpdate {
	lgr := log.LoggerFromContext(ctx)
	metadataParser := wrk.Metadata.Parser(lgr)

	dataSetID := metadataParser.String("dataSetId")
	if dataSetID == nil {
		// TODO: error
	}

	dataSet, err := i.DataRepository().GetDataSet(ctx, *dataSetID)
	if err != nil {
		// TODO: error
	} else if dataSet == nil {
		// TODO: error
	} else if dataSet.UserID == nil {
		// TODO: error
	}

	deduplicator, err := i.DataDeduplicatorFactory().Get(ctx, dataSet)
	if err != nil {
		// TODO: error
	} else if deduplicator == nil {
		// TODO: error
	}

	updatesSummary := make(map[string]struct{})

	var rawArray []*dataRaw.Raw
	if dataSet.HasDataSetTypeContinuous() {
		rawID := metadataParser.String("rawId")
		if rawID == nil {
			// TODO: error
		}

		raw, err := i.DataRawClient().Get(ctx, *rawID, nil)
		if err != nil {
			// TODO: error
		} else if raw == nil {
			// TODO: error
		}

		rawArray = append(rawArray, raw)
	} else {
		filter := dataRaw.NewFilter()
		filter.DataSetIDs = pointer.FromStringArray([]string{*dataSetID})
		pagination := page.NewPagination()

		for {
			array, err := i.DataRawClient().List(ctx, *dataSet.UserID, filter, pagination)
			if err != nil {
				// TODO: error
			} else if len(array) == 0 {
				break
			}

			rawArray = append(rawArray, array...)
		}
	}

	for _, raw := range rawArray {

		provenance := data.ParseProvenance(raw.Metadata.Parser(lgr).WithReferenceObjectParser("provenance"))

		rawContent, err := i.DataRawClient().GetContent(ctx, raw.ID, nil)
		if err != nil {
			// TODO: error
		} else if rawContent == nil {
			// TODO: error
		}
		defer rawContent.ReadCloser.Close()

		var array []interface{}
		decoder := json.NewDecoder(rawContent.ReadCloser)
		if err := decoder.Decode(&array); err != nil {
			// TODO: error
		}

		parser := structureParser.NewArray(lgr, &array)
		validator := structureValidator.New(lgr)
		normalizer := dataNormalizer.New(lgr)

		datumArray := []data.Datum{}
		for _, reference := range parser.References() {
			if datum := dataTypesFactory.ParseDatum(parser.WithReferenceObjectParser(reference)); datum != nil && *datum != nil {
				(*datum).Validate(validator.WithReference(strconv.Itoa(reference)))
				(*datum).Normalize(normalizer.WithReference(strconv.Itoa(reference)))
				datumArray = append(datumArray, *datum)
			}
		}
		parser.NotParsed()

		err = errors.Append(parser.Error(), validator.Error(), normalizer.Error())
		if err != nil {
			// TODO: error
		}

		datumArray = append(datumArray, normalizer.Data()...)
		for _, datum := range datumArray {
			datum.SetUserID(dataSet.UserID)
			datum.SetDataSetID(dataSet.UploadID)
			datum.SetProvenance(provenance)
		}

		if err = deduplicator.AddData(ctx, i.DataRepository(), dataSet, datumArray); err != nil {
			// TODO: error
		}

		for _, datum := range datumArray {
			summary.CheckDatumUpdatesSummary(updatesSummary, datum)
		}
	}

	summary.MaybeUpdateSummary(ctx, i.SummarizerRegistry(), updatesSummary, *dataSet.UserID, types.OutdatedReasonDataAdded)

	return nil
}
