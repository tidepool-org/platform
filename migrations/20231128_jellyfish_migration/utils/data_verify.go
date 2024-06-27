package utils

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

type DataVerify struct {
	ctx   context.Context
	dataC *mongo.Collection
}

func NewVerifier(ctx context.Context, dataC *mongo.Collection) (*DataVerify, error) {

	if dataC == nil {
		return nil, errors.New("missing required data collection")
	}

	m := &DataVerify{
		ctx:   ctx,
		dataC: dataC,
	}

	return m, nil
}

func (m *DataVerify) Verify(ref string, a string, b string) error {

	datasetA, err := fetchDataSet(m.ctx, m.dataC, a)
	if err != nil {
		return err
	}

	datasetB, err := fetchDataSet(m.ctx, m.dataC, b)
	if err != nil {
		return err
	}

	difference, err := CompareDatasets(datasetA, datasetB)
	if err != nil {
		return err
	}
	fmt.Printf("%v", difference)
	return nil
}
