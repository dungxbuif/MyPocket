package ratelimit

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Limiter interface {
	Allow(context.Context, string, int, time.Duration) (bool, time.Duration, error)
}
type Redis struct {
	addr, password string
	db             int
	timeout        time.Duration
}

func NewRedis(raw string) (*Redis, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "redis" {
		return nil, fmt.Errorf("invalid redis URL")
	}
	addr := parsed.Host
	if !strings.Contains(addr, ":") {
		addr += ":6379"
	}
	db := 0
	if path := strings.Trim(parsed.Path, "/"); path != "" {
		db, err = strconv.Atoi(path)
		if err != nil {
			return nil, fmt.Errorf("invalid redis database")
		}
	}
	password, _ := parsed.User.Password()
	return &Redis{addr: addr, password: password, db: db, timeout: 750 * time.Millisecond}, nil
}
func (r *Redis) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	if r == nil || limit <= 0 || window <= 0 {
		return false, 0, fmt.Errorf("rate limiter unavailable")
	}
	script := `local c=redis.call('INCR',KEYS[1]);if c==1 then redis.call('PEXPIRE',KEYS[1],ARGV[1]) end;local ttl=redis.call('PTTL',KEYS[1]);return tostring(c)..':'..tostring(ttl)`
	value, err := r.command(ctx, "EVAL", script, "1", "mypocket:rate:"+key, strconv.FormatInt(window.Milliseconds(), 10))
	if err != nil {
		return false, 0, err
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return false, 0, fmt.Errorf("invalid rate limiter response")
	}
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, 0, err
	}
	millis, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, 0, err
	}
	retry := time.Duration(millis) * time.Millisecond
	if retry < time.Second {
		retry = time.Second
	}
	return count <= limit, retry, nil
}
func (r *Redis) command(ctx context.Context, args ...string) (string, error) {
	dialer := net.Dialer{Timeout: r.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", r.addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(r.timeout))
	reader := bufio.NewReader(conn)
	if r.password != "" {
		if err = write(conn, "AUTH", r.password); err != nil {
			return "", err
		}
		if _, err = read(reader); err != nil {
			return "", err
		}
	}
	if r.db > 0 {
		if err = write(conn, "SELECT", strconv.Itoa(r.db)); err != nil {
			return "", err
		}
		if _, err = read(reader); err != nil {
			return "", err
		}
	}
	if err = write(conn, args...); err != nil {
		return "", err
	}
	return read(reader)
}
func write(w io.Writer, args ...string) error {
	var b strings.Builder
	b.WriteString("*" + strconv.Itoa(len(args)) + "\r\n")
	for _, arg := range args {
		b.WriteString("$" + strconv.Itoa(len(arg)) + "\r\n" + arg + "\r\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}
func read(r *bufio.Reader) (string, error) {
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
		n, err := strconv.Atoi(line)
		if err != nil {
			return "", err
		}
		buf := make([]byte, n+2)
		_, err = io.ReadFull(r, buf)
		return string(buf[:n]), err
	case '-':
		return "", fmt.Errorf("redis command failed")
	default:
		return "", fmt.Errorf("redis protocol error")
	}
}
