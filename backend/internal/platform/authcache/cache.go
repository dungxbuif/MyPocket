package authcache

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

type Noop struct{}

func (Noop) Get(context.Context, string) (string, bool, error) { return "", false, nil }
func (Noop) Set(context.Context, string, string, time.Duration) error {
	return nil
}
func (Noop) Delete(context.Context, string) error { return nil }
func (Noop) Close() error                         { return nil }

func DigestToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type Redis struct {
	addr     string
	password string
	db       int
	timeout  time.Duration
}

func NewRedis(rawURL string) (*Redis, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if parsed.Scheme != "redis" {
		return nil, fmt.Errorf("redis url must use redis scheme")
	}
	addr := parsed.Host
	if !strings.Contains(addr, ":") {
		addr += ":6379"
	}
	db := 0
	if path := strings.Trim(parsed.Path, "/"); path != "" {
		db, err = strconv.Atoi(path)
		if err != nil {
			return nil, fmt.Errorf("parse redis db: %w", err)
		}
	}
	password, _ := parsed.User.Password()
	return &Redis{addr: addr, password: password, db: db, timeout: 750 * time.Millisecond}, nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, bool, error) {
	resp, err := r.command(ctx, "GET", key)
	if err != nil {
		return "", false, err
	}
	if resp == "" {
		return "", false, nil
	}
	return resp, true, nil
}

func (r *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	_, err := r.command(ctx, "SETEX", key, strconv.Itoa(int(ttl.Seconds())), value)
	return err
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	_, err := r.command(ctx, "DEL", key)
	return err
}

func (r *Redis) Close() error { return nil }

func (r *Redis) command(ctx context.Context, args ...string) (string, error) {
	if r == nil {
		return "", nil
	}
	dialer := net.Dialer{Timeout: r.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", r.addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(r.timeout))
	reader := bufio.NewReader(conn)
	if r.password != "" {
		if _, err := writeRESP(conn, "AUTH", r.password); err != nil {
			return "", err
		}
		if _, err := readRESP(reader); err != nil {
			return "", err
		}
	}
	if r.db > 0 {
		if _, err := writeRESP(conn, "SELECT", strconv.Itoa(r.db)); err != nil {
			return "", err
		}
		if _, err := readRESP(reader); err != nil {
			return "", err
		}
	}
	if _, err := writeRESP(conn, args...); err != nil {
		return "", err
	}
	return readRESP(reader)
}

func writeRESP(w io.Writer, args ...string) (int, error) {
	var b strings.Builder
	b.WriteString("*")
	b.WriteString(strconv.Itoa(len(args)))
	b.WriteString("\r\n")
	for _, arg := range args {
		b.WriteString("$")
		b.WriteString(strconv.Itoa(len(arg)))
		b.WriteString("\r\n")
		b.WriteString(arg)
		b.WriteString("\r\n")
	}
	return io.WriteString(w, b.String())
}

func readRESP(r *bufio.Reader) (string, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	switch prefix {
	case '+', ':':
		return line, nil
	case '$':
		length, err := strconv.Atoi(line)
		if err != nil {
			return "", err
		}
		if length < 0 {
			return "", nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		return string(buf[:length]), nil
	case '-':
		return "", fmt.Errorf("redis error: %s", line)
	default:
		return "", fmt.Errorf("redis protocol error")
	}
}
