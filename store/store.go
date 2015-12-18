package store

type Store interface {
	Save(d interface{}) error
	Update(d interface{}) error
	Delete(id string) error
	Read(id string) (interface{}, error)
	ReadAll() ([]interface{}, error)
}

type MongoStore struct{}

func NewMongoStore() *MongoStore {
	return &MongoStore{}
}

func (this *MongoStore) Save(d interface{}) error { return nil }

func (this *MongoStore) Update(d interface{}) error { return nil }

func (this *MongoStore) Delete(id string) error { return nil }

func (this *MongoStore) Read(id string) (interface{}, error) { return nil, nil }

func (this *MongoStore) ReadAll() ([]interface{}, error) { return nil, nil }
