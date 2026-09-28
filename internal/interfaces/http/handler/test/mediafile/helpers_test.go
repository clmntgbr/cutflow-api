package mediafiletest

import (
	"bytes"
	"context"
	"errors"
	"io"

	querymediafile "go-api/internal/application/query/mediafile"
	domainmediafile "go-api/internal/domain/mediafile"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

type mockGetOwnedMediaFileHandler struct {
	called bool
	query  querymediafile.GetOwnedMediaFileQuery
	view   *domainmediafile.MediaFileThumbnailView
	err    error
}

func (m *mockGetOwnedMediaFileHandler) Handle(
	_ context.Context,
	q querymediafile.GetOwnedMediaFileQuery,
) (*domainmediafile.MediaFileThumbnailView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockThumbnailStorage struct {
	called bool
	key    string
	body   []byte
	err    error
}

func (m *mockThumbnailStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	m.called = true
	m.key = key
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.body)), nil
}

type readFailCloser struct{}

func (readFailCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (readFailCloser) Close() error             { return nil }

type mockFailingReadStorage struct{}

func (mockFailingReadStorage) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return readFailCloser{}, nil
}

type thumbnailStorage interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

func newMediaFileHandler(owned *mockGetOwnedMediaFileHandler, storage thumbnailStorage) *handler.MediaFileHandler {
	if owned == nil {
		owned = &mockGetOwnedMediaFileHandler{}
	}
	if storage == nil {
		storage = &mockThumbnailStorage{body: []byte{0xff, 0xd8, 0xff}}
	}
	return handler.NewMediaFileHandler(owned, storage)
}

func sampleThumbnailView() *domainmediafile.MediaFileThumbnailView {
	return &domainmediafile.MediaFileThumbnailView{
		ID:           testutil.TestMediaFileID,
		ThumbnailKey: "videos/" + testutil.TestMediaFileID.String() + "/thumbnail.jpg",
	}
}

func unusedUUID() uuid.UUID {
	return uuid.MustParse("01960000-0000-7000-8000-000000000099")
}

var errUnexpected = errors.New("unexpected handler call")
