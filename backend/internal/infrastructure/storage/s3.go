// Package storage provides private S3-compatible transaction attachment storage.
package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const signedGetTTL = 5 * time.Minute

type S3Config struct {
	Endpoint, Region, Bucket, Prefix, Environment, AccessKeyID, SecretAccessKey string
	ForcePathStyle                                                              bool
}

type S3 struct {
	cfg      S3Config
	endpoint *url.URL
	http     *http.Client
}

var safeFilename = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._ -]{0,240}$`)

func NewS3(cfg S3Config) (*S3, error) {
	cfg.Endpoint, cfg.Region, cfg.Bucket = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"), strings.TrimSpace(cfg.Region), strings.TrimSpace(cfg.Bucket)
	cfg.Prefix, cfg.Environment = strings.Trim(strings.TrimSpace(cfg.Prefix), "/"), strings.TrimSpace(cfg.Environment)
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.Bucket == "" || cfg.Environment == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, errors.New("S3 storage is not configured")
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid S3 endpoint")
	}
	if strings.Contains(cfg.Environment, "/") || strings.Contains(cfg.Environment, "..") {
		return nil, errors.New("invalid S3 environment")
	}
	return &S3{cfg: cfg, endpoint: u, http: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (s *S3) Key(owner, batch, attachment, filename string) (string, error) {
	for _, part := range []string{owner, batch, attachment} {
		if part == "" || strings.ContainsAny(part, "/\\") || strings.Contains(part, "..") {
			return "", errors.New("invalid attachment key identifier")
		}
	}
	if !safeFilename.MatchString(filename) || strings.Contains(filename, "..") {
		return "", errors.New("invalid attachment filename")
	}
	name := strings.ReplaceAll(filename, " ", "-")
	parts := []string{s.cfg.Prefix, s.cfg.Environment, "owners", owner, "batches", batch, attachment, name}
	return strings.Trim(strings.Join(parts, "/"), "/"), nil
}

func (s *S3) Put(ctx context.Context, key, contentType string, data []byte) error {
	if key == "" || len(data) == 0 || contentType == "" {
		return errors.New("invalid attachment upload")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.objectURL(key).String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprint(len(data)))
	if err := s.sign(req, hex.EncodeToString(hash(data)), time.Now()); err != nil {
		return err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return errors.New("S3 upload failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("S3 upload returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.objectURL(key).String(), nil)
	if err != nil {
		return err
	}
	if err := s.sign(req, hex.EncodeToString(hash(nil)), time.Now()); err != nil {
		return err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return errors.New("S3 delete failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound && (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		return fmt.Errorf("S3 delete returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *S3) SignedGet(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if key == "" {
		return "", errors.New("invalid attachment key")
	}
	u := s.objectURL(key)
	now := time.Now().UTC()
	scopeDate := now.Format("20060102")
	scope := scopeDate + "/" + s.cfg.Region + "/s3/aws4_request"
	q := u.Query()
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", s.cfg.AccessKeyID+"/"+scope)
	q.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	q.Set("X-Amz-Expires", fmt.Sprint(int(signedGetTTL.Seconds())))
	q.Set("X-Amz-SignedHeaders", "host")
	u.RawQuery = q.Encode()
	canonical := strings.Join([]string{http.MethodGet, canonicalURI(u), canonicalQuery(u.Query()), "host:" + u.Host + "\n", "host", "UNSIGNED-PAYLOAD"}, "\n")
	toSign := strings.Join([]string{"AWS4-HMAC-SHA256", now.Format("20060102T150405Z"), scope, hex.EncodeToString(hash([]byte(canonical)))}, "\n")
	q.Set("X-Amz-Signature", hex.EncodeToString(hmacSHA256(s.signingKey(scopeDate), []byte(toSign))))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *S3) objectURL(key string) *url.URL {
	u := *s.endpoint
	if s.cfg.ForcePathStyle {
		u.Path = path.Join(u.Path, s.cfg.Bucket, key)
	} else {
		u.Host = s.cfg.Bucket + "." + u.Host
		u.Path = path.Join(u.Path, key)
	}
	return &u
}
func (s *S3) sign(req *http.Request, payload string, now time.Time) error {
	now = now.UTC()
	scopeDate := now.Format("20060102")
	scope := scopeDate + "/" + s.cfg.Region + "/s3/aws4_request"
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Amz-Content-Sha256", payload)
	req.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	canonicalHeaders := "host:" + req.URL.Host + "\n" + "x-amz-content-sha256:" + payload + "\n" + "x-amz-date:" + now.Format("20060102T150405Z") + "\n"
	signed := "host;x-amz-content-sha256;x-amz-date"
	canonical := strings.Join([]string{req.Method, canonicalURI(req.URL), canonicalQuery(req.URL.Query()), canonicalHeaders, signed, payload}, "\n")
	toSign := strings.Join([]string{"AWS4-HMAC-SHA256", now.Format("20060102T150405Z"), scope, hex.EncodeToString(hash([]byte(canonical)))}, "\n")
	signature := hex.EncodeToString(hmacSHA256(s.signingKey(scopeDate), []byte(toSign)))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.cfg.AccessKeyID+"/"+scope+", SignedHeaders="+signed+", Signature="+signature)
	return nil
}
func (s *S3) signingKey(date string) []byte {
	k := hmacSHA256([]byte("AWS4"+s.cfg.SecretAccessKey), []byte(date))
	k = hmacSHA256(k, []byte(s.cfg.Region))
	k = hmacSHA256(k, []byte("s3"))
	return hmacSHA256(k, []byte("aws4_request"))
}
func canonicalURI(u *url.URL) string     { return (&url.URL{Path: u.EscapedPath()}).EscapedPath() }
func canonicalQuery(q url.Values) string { return q.Encode() }
func hash(data []byte) []byte            { sum := sha256.Sum256(data); return sum[:] }
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = io.WriteString(h, string(data))
	return h.Sum(nil)
}
