package api

type Api interface {
	ValidateStruct(any interface{}) (err error)
}

type CeleryInterface interface {
	Data() (data map[string]interface{}, err error)
}

type ModelCurd interface {
	Update(data map[string]interface{}) (err error)
}

type RecordWebsocketLog interface {
	RecordLog(data map[string]interface{}) error
}
