package api

type Api interface {
	ValidateStruct(any interface{}) (err error)
}

type CeleryInterface interface {
	Data() (data map[string]interface{}, err error)
}
