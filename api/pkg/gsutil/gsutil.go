package gsutil

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bboughton/gcp-helpers/gsurl"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

// parseURL exists to allow us to use 'url' as a request param
var parseURL = url.Parse

type GSUtil interface {
	ParseURL(url string) (*URL, error)
	ReadFile(ctx context.Context, url string) ([]byte, error)
	WriteFile(ctx context.Context, url, content string) error
}

type gsutil struct {
	client *storage.Client
}

// NewClient initializes GSUtil client for working with Google Cloud Storage.
func NewClient(ctx context.Context) (GSUtil, error) {
	client, err := storage.NewClient(ctx, option.WithScopes(storage.ScopeReadOnly))
	if err != nil {
		return nil, err
	}

	return &gsutil{
		client: client,
	}, nil
}

func (g *gsutil) Close() error {
	return g.client.Close()
}

// ReadFile reads the contents of a file located in GCloud Storage returning
// a byte array with its contents or an error. The url is expected to have the
// following format "gs://the-bucket-name/path/to/the/file.ext". Users of this
// func should take care to only load objects of a known size as errors can
// arise if the requested object is too large.
func (g *gsutil) ReadFile(ctx context.Context, url string) ([]byte, error) {
	u, err := gsurl.Parse(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url %s", url)
	}

	reader, err := g.client.Bucket(u.Bucket).Object(u.Object).NewReader(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file from %s", url)
	}
	defer reader.Close() //nolint:errcheck

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read the content of %s", url)
	}
	return bytes, nil
}

// URL contains the information needed to identify the location of an object
// located in Google Cloud Storage.
type URL struct {
	// Bucket is the name of the Google Cloud Storage bucket where the object
	// is located.
	Bucket string

	// Object is the name and or path of the object stored in the bucket. It
	// should not start with a foward slash.
	Object string
}

// Parse parses a Google Cloud Storage string into a URL struct. The expected
// format of the string is gs://[bucket-name]/[object-path]. If the provided
// URL is formatted incorrectly an error will be returned.
func (g *gsutil) ParseURL(url string) (*URL, error) {
	u, err := parseURL(url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url %s", url)
	}
	if u.Scheme != "gs" {
		return nil, errors.New("invalid protocal specified, the only protocal that is permitted is 'gs'")
	}

	bucket, object := u.Host, strings.TrimLeft(u.Path, "/")

	if bucket == "" {
		return nil, errors.New("bucket name is required")
	}

	if object == "" {
		return nil, errors.New("object name is required")
	}

	return &URL{
		Bucket: bucket,
		Object: object,
	}, nil
}

func (g *gsutil) WriteFile(ctx context.Context, url, content string) error {
	u, err := gsurl.Parse(url)
	if err != nil {
		return errors.Wrapf(err, "failed to parse url %s", url)
	}
	w := g.client.Bucket(u.Bucket).Object(u.Object).NewWriter(ctx)

	if _, err := fmt.Fprint(w, content); err != nil {
		return errors.Wrapf(err, "failed to write content to %s", url)
	}

	if err := w.Close(); err != nil {
		return errors.Wrapf(err, "failed to close writer for %s", url)
	}

	return nil
}
