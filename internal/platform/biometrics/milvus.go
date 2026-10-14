package biometrics

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusBridge struct {
	client client.Client
}

func NewMilvusBridge(ctx context.Context, addr string) (*MilvusBridge, error) {
	c, err := client.NewClient(ctx, client.Config{
		Address: addr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %w", err)
	}

	return &MilvusBridge{client: c}, nil
}

func (b *MilvusBridge) CreateBiometricCollection(ctx context.Context, name string, dimension int) error {
	schema := &entity.Schema{
		CollectionName: name,
		Description:    "SNISID Biometric Vector Store",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "identity_id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				TypeParams: map[string]string{"max_length": "128"},
			},
			{
				Name:     "biometric_vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", dimension),
				},
			},
		},
	}

	return b.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
}

func (b *MilvusBridge) InsertBiometric(ctx context.Context, collection, id string, vector []float32) error {
	idCol := entity.NewColumnVarChar("identity_id", []string{id})
	vecCol := entity.NewColumnFloatVector("biometric_vector", int(len(vector)), [][]float32{vector})

	_, err := b.client.Insert(ctx, collection, "", idCol, vecCol)
	return err
}

func (b *MilvusBridge) Search(ctx context.Context, collection string, vector []float32) (string, float32, error) {
	searchParam, _ := entity.NewIndexIvfFlatSearchParam(10)

	results, err := b.client.Search(ctx, collection, []string{}, "", []string{"identity_id"},
		[]entity.Vector{entity.FloatVector(vector)}, "biometric_vector", entity.L2, 1, searchParam)

	if err != nil {
		return "", 0, err
	}

	if len(results) == 0 || results[0].ResultCount == 0 {
		return "", 0, fmt.Errorf("no match found")
	}

	matchID, _ := results[0].Fields.GetColumn("identity_id").GetAsString(0)
	distance := results[0].Scores[0]

	return matchID, distance, nil
}
