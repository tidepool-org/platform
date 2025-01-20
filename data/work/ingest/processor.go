package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data"
	dataDeduplicator "github.com/tidepool-org/platform/data/deduplicator"
	dataNormalizer "github.com/tidepool-org/platform/data/normalizer"
	dataRaw "github.com/tidepool-org/platform/data/raw"
	dataStore "github.com/tidepool-org/platform/data/store"
	dataSummary "github.com/tidepool-org/platform/data/summary"
	dataSummaryTypes "github.com/tidepool-org/platform/data/summary/types"
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
	Namespace           = "org.tidepool.platform.data.work"
	Type                = "org.tidepool.platform.data.work.ingest"
	ProcessingTimeout   = 86400
	RetryDuration       = 5 * time.Second
	RetryDurationJitter = 1 * time.Second
)

func NewCreateForDataSetTypeNormal(dataSet *data.DataSet) (*work.Create, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	} else if !dataSet.HasDataSetTypeNormal() {
		return nil, errors.New("data set type is not normal")
	}

	namespacedDataSetID := pointer.FromString(fmt.Sprintf("%s:%s", Namespace, *dataSet.ID))

	metadata := metadata.NewMetadata()
	metadata.Set("dataSetId", dataSet.ID)

	return &work.Create{
		Type:              Type,
		GroupID:           namespacedDataSetID,
		DeduplicationID:   namespacedDataSetID,
		ProcessingTimeout: ProcessingTimeout,
		Metadata:          metadata,
	}, nil
}

func NewCreateForDataSetTypeContinuous(dataSet *data.DataSet, raw *dataRaw.Raw) (*work.Create, error) {
	if dataSet == nil {
		return nil, errors.New("data set is missing")
	} else if !dataSet.HasDataSetTypeContinuous() {
		return nil, errors.New("data set type is not continuous")
	}
	if raw == nil {
		return nil, errors.New("raw is missing")
	}

	namespacedDataSetID := pointer.FromString(fmt.Sprintf("%s:%s", Namespace, *dataSet.ID))
	namespacedRawID := pointer.FromString(fmt.Sprintf("%s:%s", *namespacedDataSetID, raw.ID))

	metadata := metadata.NewMetadata()
	metadata.Set("dataSetId", dataSet.ID)
	metadata.Set("rawId", raw.ID)

	return &work.Create{
		Type:              Type,
		GroupID:           namespacedDataSetID,
		DeduplicationID:   namespacedRawID,
		SerialID:          namespacedDataSetID,
		ProcessingTimeout: ProcessingTimeout,
		Metadata:          metadata,
	}, nil
}

type DataRawClient interface {
	List(ctx context.Context, userID string, filter *dataRaw.Filter, pagination *page.Pagination) ([]*dataRaw.Raw, error)
	Get(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Raw, error)
	GetContent(ctx context.Context, id string, condition *request.Condition) (*dataRaw.Content, error)
	Update(ctx context.Context, id string, condition *request.Condition, update *dataRaw.Update) (*dataRaw.Raw, error)
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

func (p *Processor) Process(ctx context.Context, wrk *work.Work, updater work.ProcessingUpdater) work.ProcessResult {
	processor := &processor{
		Processor: p,
		context:   ctx,
		work:      wrk,
		updater:   updater,
	}
	return processor.Process()
}

type processor struct {
	*Processor
	context          context.Context
	work             *work.Work
	updater          work.ProcessingUpdater
	dataSet          *data.DataSet
	dataDeduplicator dataDeduplicator.Deduplicator
	rawArray         []*dataRaw.Raw
}

func (p *processor) Process() work.ProcessResult {
	processes := []func() *work.ProcessResult{
		p.getDataSet,
		p.getDataDeduplicator,
		p.getRawArray,
		p.processRawArray,
		p.finalize,
	}
	for _, process := range processes {
		if processResult := process(); processResult != nil {
			return *processResult
		}
	}
	return *p.done()
}

func (p *processor) getDataSet() *work.ProcessResult {
	dataSetID := p.work.Metadata.Parser(log.LoggerFromContext(p.context)).String("dataSetId")
	if dataSetID == nil {
		return p.failed(errors.New("data set id is missing"))
	}

	dataSet, err := p.DataRepository.GetDataSet(p.context, *dataSetID)
	if err != nil {
		return p.failing(err)
	} else if dataSet == nil {
		return p.failed(errors.New("data set is missing"))
	} else if dataSet.UserID == nil {
		return p.failed(errors.New("data set user id is missing"))
	}

	p.dataSet = dataSet
	return nil
}

func (p *processor) getDataDeduplicator() *work.ProcessResult {
	dataDeduplicator, err := p.DataDeduplicatorFactory.Get(p.context, p.dataSet)
	if err != nil {
		return p.failed(err)
	} else if dataDeduplicator == nil {
		return p.failed(errors.New("data deduplicator does not exist for data set"))
	}

	p.dataDeduplicator = dataDeduplicator
	return nil
}

func (p *processor) getRawArray() *work.ProcessResult {
	if p.dataSet.HasDataSetTypeNormal() {
		return p.getRawArrayForDataSetTypeNormal()
	} else if p.dataSet.HasDataSetTypeContinuous() {
		return p.getRawArrayForDataSetTypeContinuous()
	} else {
		return p.failed(errors.New("unknown data set type"))
	}
}

func (p *processor) getRawArrayForDataSetTypeNormal() *work.ProcessResult {
	var rawArray []*dataRaw.Raw

	filter := &dataRaw.Filter{
		DataSetIDs: pointer.FromStringArray([]string{*p.dataSet.ID}),
		Processed:  pointer.FromBool(false),
	}
	pagination := page.NewPagination()

	for {
		array, err := p.DataRawClient.List(p.context, *p.dataSet.UserID, filter, pagination)
		if err != nil {
			return p.failing(err)
		}
		rawArray = append(rawArray, array...)
		if len(array) < pagination.Size {
			break
		}
		pagination.Page += 1
	}

	p.rawArray = rawArray
	return nil
}

func (p *processor) getRawArrayForDataSetTypeContinuous() *work.ProcessResult {
	var rawArray []*dataRaw.Raw

	rawID := p.work.Metadata.Parser(log.LoggerFromContext(p.context)).String("rawId")
	if rawID == nil {
		return p.failed(errors.New("raw id is missing"))
	}

	raw, err := p.DataRawClient.Get(p.context, *rawID, nil)
	if err != nil {
		return p.failing(err)
	} else if raw == nil {
		return p.failed(errors.New("raw is missing"))
	}

	if !raw.Processed() {
		rawArray = append(rawArray, raw)
	}

	p.rawArray = rawArray
	return nil
}

func (p *processor) processRawArray() *work.ProcessResult {
	if len(p.rawArray) < 1 {
		log.LoggerFromContext(p.context).Warn("No raw data to process for data set")
		return p.done()
	}

	updatesSummary := make(map[string]struct{})
	defer dataSummary.MaybeUpdateSummary(context.WithoutCancel(p.context), p.SummarizerRegistry, updatesSummary, *p.dataSet.UserID, dataSummaryTypes.OutdatedReasonDataAdded)

	for _, raw := range p.rawArray {
		lgr := log.LoggerFromContext(p.context).WithField("raw", raw)
		provenance := data.ParseProvenance(raw.Metadata.Parser(lgr).WithReferenceObjectParser("provenance"))

		condition := &request.Condition{Revision: &raw.Revision}
		rawContent, err := p.DataRawClient.GetContent(p.context, raw.ID, condition)
		if err != nil {
			return p.failing(err)
		} else if rawContent == nil {
			lgr.Warn("No content for raw")
			continue
		}
		defer rawContent.ReadCloser.Close()

		var array []interface{}
		decoder := json.NewDecoder(rawContent.ReadCloser)
		if err := decoder.Decode(&array); err != nil {
			// TODO: raw data error
			continue
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
			// TODO: raw data error
			continue
		}

		datumArray = append(datumArray, normalizer.Data()...)
		for _, datum := range datumArray {
			datum.SetUserID(p.dataSet.UserID)
			datum.SetDataSetID(p.dataSet.UploadID)
			datum.SetProvenance(provenance)
		}

		if err = p.dataDeduplicator.AddData(p.context, p.DataRepository, p.dataSet, datumArray); err != nil {
			return p.failing(err)
		}

		for _, datum := range datumArray {
			dataSummary.CheckDatumUpdatesSummary(updatesSummary, datum)
		}

		update := &dataRaw.Update{ProcessedTime: time.Now()}
		if raw, err := p.DataRawClient.Update(p.context, raw.ID, condition, update); err != nil {
			return p.failing(err)
		} else if raw == nil {
			lgr.Warn("Raw not updated with processed time")
		}
	}

	return nil
}

func (p *processor) finalize() *work.ProcessResult {
	if p.dataSet.HasDataSetTypeNormal() {
		if err := p.dataDeduplicator.Close(p.context, p.DataRepository, p.dataSet); err != nil {
			return p.failing(err)
		}

		updatesSummary := make(map[string]struct{})
		for _, typ := range dataSummaryTypes.AllSummaryTypes {
			updatesSummary[typ] = struct{}{}
		}
		dataSummary.MaybeUpdateSummary(context.WithoutCancel(p.context), p.SummarizerRegistry, updatesSummary, *p.dataSet.UserID, dataSummaryTypes.OutdatedReasonUploadCompleted)
	}

	return nil
}

func (p *processor) failing(err error) *work.ProcessResult {
	failingRetryCount := *pointer.DefaultInt(p.work.FailingRetryCount, 0) + 1
	processResult := work.NewProcessResultFailing(work.FailingUpdate{
		FailingError:      errors.Serializable{Error: err},
		FailingRetryCount: failingRetryCount,
		FailingRetryTime:  time.Now().Add(retryDuration(failingRetryCount)),
		Metadata:          p.work.Metadata,
	})
	return &processResult
}

func (p *processor) failed(err error) *work.ProcessResult {
	processResult := work.NewProcessResultFailed(work.FailedUpdate{
		FailedError: errors.Serializable{Error: err},
		Metadata:    p.work.Metadata,
	})
	return &processResult
}

func (p *processor) done() *work.ProcessResult {
	processResult := work.NewProcessResultDelete()
	return &processResult
}

func retryDuration(retryCount int) time.Duration {
	fallbackFactor := time.Duration(1 << (retryCount - 1))
	retryDurationJitter := int64(RetryDurationJitter * fallbackFactor)
	return RetryDuration*fallbackFactor + time.Duration(rand.Int63n(2*retryDurationJitter)-retryDurationJitter)
}
