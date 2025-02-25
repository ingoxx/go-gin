package assets

import (
	"github.com/ingoxx/go-gin/project/model"
)

type ProgramOperate struct {
}

func (po *ProgramOperate) ProgramData(al model.AssetsProgramModel) (map[string]interface{}, error) {
	var pn = make(map[string]interface{})
	var data = make(map[string]interface{})

	pl, err := al.List()
	if err != nil {
		return nil, err
	}

	for _, v := range pl {
		pn[v.CnName] = v.EnName
	}

	data["pl"] = pl
	data["pn"] = pn

	return data, nil
}

func NewProgramOperate() *ProgramOperate {
	return &ProgramOperate{}
}
