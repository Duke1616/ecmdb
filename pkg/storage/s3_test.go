package storage

import (
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestBuildObjectKey(t *testing.T) {
	fixedTime := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		prefix   string
		fileName string
		wantKey  string
	}{
		{
			name:     "standard prefix and filename",
			prefix:   "import",
			fileName: "report.xlsx",
			wantKey:  "import/2026-09-04/report.xlsx",
		},
		{
			name:     "prefix with leading and trailing slashes",
			prefix:   "/import/template/",
			fileName: "users.xlsx",
			wantKey:  "import/template/2026-09-04/users.xlsx",
		},
		{
			name:     "empty prefix",
			prefix:   "",
			fileName: "export.xlsx",
			wantKey:  "2026-09-04/export.xlsx",
		},
		{
			name:     "only slashes in prefix",
			prefix:   "///",
			fileName: "data.csv",
			wantKey:  "2026-09-04/data.csv",
		},
		{
			name:     "multi-level filename",
			prefix:   "cmdb",
			fileName: "sub/dir/asset.xlsx",
			wantKey:  "cmdb/2026-09-04/sub/dir/asset.xlsx",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := buildObjectKey(tc.prefix, tc.fileName, fixedTime)
			assert.Equal(t, tc.wantKey, got)
		})
	}
}

func TestNewS3Storage_ImplementsInterface(t *testing.T) {
	var _ IStorage = (*s3Storage)(nil)

	s := NewS3Storage(&minio.Client{})
	assert.NotNil(t, s)
}
