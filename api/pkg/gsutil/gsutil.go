package gsutil

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// parseURL exists to allow us to use 'url' as a request param
// var parseURL = url.Parse

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

type GSUtil interface {
	ParseURL(url string) (*URL, error)
	ReadFile(ctx context.Context, url string) ([]byte, error)
	WriteFile(ctx context.Context, url, content string) error
}

type gsutil struct {
	client *storage.Client
}

type Option struct {
	AuthenticationType string
	CredentialsJson    string
}

// NewClient initializes GSUtil client for working with Google Cloud Storage.
func NewClient(ctx context.Context, opt Option) (GSUtil, error) {
	clientOpts := []option.ClientOption{}
	clientOpts = append(clientOpts, option.WithScopes(storage.ScopeReadWrite))

	if opt.AuthenticationType == "" {
		clientOpts = append(clientOpts, option.WithoutAuthentication())
	} else if opt.AuthenticationType == "credentials_json" {
		credsByte, err := base64.StdEncoding.DecodeString(opt.CredentialsJson)
		if err != nil {
			return nil, err
		}
		clientOpts = append(clientOpts, option.WithCredentialsJSON(credsByte))
	}

	client, err := storage.NewClient(ctx, clientOpts...)
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
	u, err := g.ParseURL(url)
	if err != nil {
		return nil, err
	}

	reader, err := g.client.Bucket(u.Bucket).Object(u.Object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close() //nolint:errcheck

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Parse parses a Google Cloud Storage string into a URL struct. The expected
// format of the string is gs://[bucket-name]/[object-path]. If the provided
// URL is formatted incorrectly an error will be returned.
func (g *gsutil) ParseURL(gsURL string) (*URL, error) {
	u, err := url.Parse(gsURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "gs" {
		return nil, err
	}

	bucket, object := u.Host, strings.TrimLeft(u.Path, "/")

	if bucket == "" {
		return nil, err
	}

	if object == "" {
		return nil, err
	}

	return &URL{
		Bucket: bucket,
		Object: object,
	}, nil
}

func (g *gsutil) WriteFile(ctx context.Context, url, content string) error {
	u, err := g.ParseURL(url)
	if err != nil {
		return err
	}
	w := g.client.Bucket(u.Bucket).Object(u.Object).NewWriter(ctx)

	if _, err := fmt.Fprint(w, content); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
