package nin

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var deptCodeMap = map[string]string{
	"OU": "OU",
	"ND": "ND",
	"NE": "NE",
	"NO": "NO",
	"AR": "AR",
	"CE": "CE",
	"SU": "SU",
	"SE": "SE",
	"GA": "GA",
	"NI": "NI",
}

var validDeptCodes []string

func init() {
	for k := range deptCodeMap {
		validDeptCodes = append(validDeptCodes, k)
	}
}

type Generator struct {
	db  *sql.DB
	mu  sync.Mutex
	seq map[string]int
}

func NewGenerator(db *sql.DB) *Generator {
	return &Generator{
		db:  db,
		seq: make(map[string]int),
	}
}

func (g *Generator) Generate(ctx context.Context, deptCode string, birthYear int) (string, error) {
	if _, ok := deptCodeMap[deptCode]; !ok {
		return "", fmt.Errorf("invalid department code: %s", deptCode)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	year := birthYear % 100
	key := fmt.Sprintf("%s-%04d", deptCode, birthYear)

	g.seq[key]++
	seq := g.seq[key]

	seqStr := fmt.Sprintf("%06d", seq)

	nin := fmt.Sprintf("HTI-%s-%02d-%s", deptCode, year, seqStr)

	return nin, nil
}

func (g *Generator) GenerateWithDB(ctx context.Context, deptCode string, birthYear int) (string, error) {
	if _, ok := deptCodeMap[deptCode]; !ok {
		return "", fmt.Errorf("invalid department code: %s", deptCode)
	}

	now := time.Now()
	yearSuffix := birthYear % 100
	seq, err := g.nextSequence(ctx, deptCode, birthYear)
	if err != nil {
		return "", err
	}
	_ = now

	return fmt.Sprintf("HTI-%s-%02d-%06d", deptCode, yearSuffix, seq), nil
}

func (g *Generator) nextSequence(ctx context.Context, deptCode string, birthYear int) (int, error) {
	var maxSeq sql.NullInt64
	err := g.db.QueryRowContext(ctx,
		`SELECT MAX(CAST(SUBSTRING(nin FROM 13 FOR 6) AS INTEGER)) FROM citizens WHERE dept_code = $1 AND EXTRACT(YEAR FROM dob) = $2`,
		deptCode, birthYear,
	).Scan(&maxSeq)
	if err != nil {
		return 0, err
	}
	if maxSeq.Valid {
		return int(maxSeq.Int64) + 1, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return 1, nil
	}
	return int(n.Int64()) + 1, nil
}

func ValidateNIN(nin string) bool {
	if len(nin) != 13 {
		return false
	}
	if nin[:4] != "HTI-" {
		return false
	}
	deptCode := nin[4:6]
	if _, ok := deptCodeMap[deptCode]; !ok {
		return false
	}
	if nin[6] != '-' {
		return false
	}
	return true
}

func ExtractDeptCode(nin string) string {
	if len(nin) >= 6 {
		return nin[4:6]
	}
	return ""
}
