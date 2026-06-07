package swagger

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

func unmarshalSpec(data []byte) (*v3.Document, error) {
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, err
	}
	model, err := document.BuildV3Model()
	if err != nil {
		return nil, err
	}
	return &model.Model, nil
}
