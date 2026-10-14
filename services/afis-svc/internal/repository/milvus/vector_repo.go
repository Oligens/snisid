package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/snisid/platform/services/afis-svc/internal/domain"
)

type VectorRepo struct {
	client     client.Client
	collection string
	dim        int
}

func NewVectorRepo(c client.Client, collection string, dim int) *VectorRepo {
	return &VectorRepo{
		client:     c,
		collection: collection,
		dim:        dim,
	}
}

type Candidate struct {
	PrintID        string
	SubjectID      string
	Score          float64
	NationalAFISID string
}

func (r *VectorRepo) InsertVectors(ctx context.Context, printIDs []string, vectors [][]float32) error {
	idCol := entity.NewColumnVarChar("print_id", printIDs)
	vecCol := entity.NewColumnFloatVector("embedding", r.dim, vectors)
	if _, err := r.client.Insert(ctx, r.collection, "", idCol, vecCol); err != nil {
		return fmt.Errorf("milvus insert: %w", err)
	}
	return nil
}

func (r *VectorRepo) SearchNearest(ctx context.Context, queryVectors [][]float32, topK int) ([]Candidate, error) {
	searchResult, err := r.client.Search(
		ctx, r.collection, nil, "",
		[]string{"print_id", "subject_id", "national_afis_id"},
		queryVectors,
		"embedding",
		entity.L2,
		topK,
	)
	if err != nil {
		return nil, fmt.Errorf("milvus search: %w", err)
	}

	var candidates []Candidate
	for _, result := range searchResult {
		for i := 0; i < result.ResultCount; i++ {
			c := Candidate{
				Score: float64(result.Scores[i]),
			}
			if id, ok := result.Fields.GetColumn("print_id"); ok {
				if col, ok := id.(*entity.ColumnVarChar); ok && col.Len() > i {
					c.PrintID = col.Data()[i]
				}
			}
			if sid, ok := result.Fields.GetColumn("subject_id"); ok {
				if col, ok := sid.(*entity.ColumnVarChar); ok && col.Len() > i {
					c.SubjectID = col.Data()[i]
				}
			}
			if nid, ok := result.Fields.GetColumn("national_afis_id"); ok {
				if col, ok := nid.(*entity.ColumnVarChar); ok && col.Len() > i {
					c.NationalAFISID = col.Data()[i]
				}
			}
			candidates = append(candidates, c)
		}
	}
	return candidates, nil
}

func (r *VectorRepo) EnsureCollection(ctx context.Context) error {
	exists, err := r.client.HasCollection(ctx, r.collection)
	if err != nil {
		return fmt.Errorf("check collection: %w", err)
	}
	if exists {
		return nil
	}

	schema := &entity.Schema{
		CollectionName: r.collection,
		AutoID:         false,
		Fields: []*entity.Field{
			{Name: "print_id", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "36"}, PrimaryKey: true},
			{Name: "subject_id", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "36"}},
			{Name: "national_afis_id", DataType: entity.FieldTypeVarChar, TypeParams: map[string]string{"max_length": "20"}},
			{Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": fmt.Sprintf("%d", r.dim)}},
		},
	}
	if err := r.client.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	idx, err := entity.NewIndexIvfFlat(entity.L2, 128)
	if err != nil {
		return fmt.Errorf("new index: %w", err)
	}
	if err := r.client.CreateIndex(ctx, r.collection, "embedding", idx, false); err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	return nil
}
