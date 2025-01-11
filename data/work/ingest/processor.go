package ingest

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataSummary "github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/data/summary/types"
	dataTypesFactory "github.com/tidepool-org/platform/data/types/factory"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/metadata"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/work"
)

const (
	Type              = "org.tidepool.platform.work.data.ingest"
	ProcessingTimeout = 300
)

func NewCreate(dataSet *data.DataSet, raw *dataRaw.Raw) *work.Create {
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

type DataRawClient interface {
	List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error)
	Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error)
	GetContent(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Content, error)
}

type DataDeduplicatorFactory interface {
	Get(ctx context.Context, dataSet *data.DataSet) (dataDeduplicator.Deduplicator, error)
}

type ProcessorDependencies struct {
	DataRawClient           DataRawClient
	DataRepository          dataStore.DataRepository
	DataDeduplicatorFactory DataDeduplicatorFactory
	SummarizerRegistry      *dataSummary.SummarizerRegistry
}

func (p ProcessorDependencies) Validate() error {
	if p.DataRawClient == nil {
		return errors.New("data raw client is missing")
	}
	if p.DataRepository == nil {
		return errors.New("data repository is missing")
	}
	if p.DataDeduplicatorFactory == nil {
		return errors.New("data deduplicator factory is missing")
	}
	if p.SummarizerRegistry == nil {
		return errors.New("summarizer registry is missing")
	}
	return nil
}

func NewProcessor(processorDependencies ProcessorDependencies) (*Processor, error) {
	if err := processorDependencies.Validate(); err != nil {
		return nil, err
	}

	return &Processor{
		ProcessorDependencies: processorDependencies,
	}, nil
}

type Processor struct {
	ProcessorDependencies
}

func (p *Processor) Type() string {
	return Type
}

func (p *Processor) Quantity() int {
	return 8 // TODO: Configuration?
}

func (p *Processor) Frequency() time.Duration {
	return 5 * time.Second // TODO: Configuration?
}

func (p *Processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) *work.PendingUpdate {
	lgr := log.LoggerFromContext(ctx)
	metadataParser := wrk.Metadata.Parser(lgr)

	dataSetID := metadataParser.String("dataSetId")
	if dataSetID == nil {
		// TODO: error
	}

	dataSet, err := p.DataRepository.GetDataSet(ctx, *dataSetID)
	if err != nil {
		// TODO: error
	} else if dataSet == nil {
		// TODO: error
	} else if dataSet.UserID == nil {
		// TODO: error
	}

	deduplicator, err := p.DataDeduplicatorFactory.Get(ctx, dataSet)
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

		raw, err := p.DataRawClient.Get(ctx, *rawID, nil)
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
			array, err := p.DataRawClient.List(ctx, *dataSet.UserID, filter, pagination)
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

		rawContent, err := p.DataRawClient.GetContent(ctx, raw.ID, nil)
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

		if err = deduplicator.AddData(ctx, p.DataRepository, dataSet, datumArray); err != nil {
			// TODO: error
		}

		for _, datum := range datumArray {
			dataSummary.CheckDatumUpdatesSummary(updatesSummary, datum)
		}
	}

	dataSummary.MaybeUpdateSummary(ctx, p.SummarizerRegistry, updatesSummary, *dataSet.UserID, types.OutdatedReasonDataAdded)

	return nil
}
