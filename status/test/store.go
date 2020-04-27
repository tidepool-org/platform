package test

func NewStoreStatusReporter() *TestStoreStatusReporter {
	return &TestStoreStatusReporter{}
}

type StoreStatus struct {
	Ping string
}

func OkStoreStatus() interface{} {
	return &StoreStatus{Ping:"OK"}
}

type TestStoreStatusReporter struct {
	sts interface{}
}

func (r *TestStoreStatusReporter) SetStatus(sts interface{}) {
	r.sts = sts
}


func (r *TestStoreStatusReporter) Status() interface{} {
	return r.sts
}
