package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/payloadmsg"
	"github.com/Wei-Shaw/sub2api/internal/repository/payloadlog"
)

var ErrPayloadLogStoreUnavailable = errors.New("payload log store unavailable")

type PayloadLogJob struct {
	Log                     *payloadlog.PayloadLog
	RequestRaw, ResponseRaw []byte
}

type PayloadLogStatus struct {
	Enabled        bool   `json:"enabled"`
	StoreAvailable bool   `json:"store_available"`
	StoreError     string `json:"store_error"`
	Dropped        uint64 `json:"dropped"`
}

type PayloadLogService struct {
	cfg       config.PayloadLogConfig
	store     payloadlog.Store
	settings  SettingRepository
	enabled   atomic.Bool
	available atomic.Bool
	dropped   atomic.Uint64
	storeErr  atomic.Value // string
	queue     chan PayloadLogJob
	stopped   atomic.Bool
	wg        sync.WaitGroup
}

func NewPayloadLogService(cfg config.PayloadLogConfig, store payloadlog.Store, settings SettingRepository) *PayloadLogService {
	s := &PayloadLogService{cfg: cfg, store: store, settings: settings}
	s.storeErr.Store("")
	if !cfg.Enabled {
		return s
	}
	s.queue = make(chan PayloadLogJob, cfg.QueueSize)
	if store != nil {
		s.available.Store(true)
	} else {
		s.storeErr.Store(ErrPayloadLogStoreUnavailable.Error())
	}
	if settings != nil {
		if v, err := settings.GetValue(context.Background(), SettingKeyPayloadLogEnabled); err == nil {
			on, _ := strconv.ParseBool(strings.TrimSpace(v))
			s.enabled.Store(on && s.available.Load())
		}
	}
	for i := 0; i < cfg.WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}
	return s
}

func (s *PayloadLogService) FeatureEnabled() bool { return s != nil && s.cfg.Enabled }
func (s *PayloadLogService) ShouldCapture() bool {
	return s != nil && s.cfg.Enabled && s.enabled.Load() && s.available.Load() && !s.stopped.Load()
}
func (s *PayloadLogService) Status() PayloadLogStatus {
	if s == nil {
		return PayloadLogStatus{}
	}
	e, _ := s.storeErr.Load().(string)
	return PayloadLogStatus{Enabled: s.enabled.Load(), StoreAvailable: s.available.Load(), StoreError: e, Dropped: s.dropped.Load()}
}
func (s *PayloadLogService) SetEnabled(ctx context.Context, on bool) error {
	if s == nil || !s.cfg.Enabled {
		return ErrPayloadLogStoreUnavailable
	}
	if on && !s.available.Load() {
		return ErrPayloadLogStoreUnavailable
	}
	if s.settings != nil {
		if err := s.settings.Set(ctx, SettingKeyPayloadLogEnabled, strconv.FormatBool(on)); err != nil {
			return err
		}
	}
	s.enabled.Store(on)
	return nil
}
func (s *PayloadLogService) TryEnqueue(j PayloadLogJob) bool {
	if !s.ShouldCapture() {
		return false
	}
	select {
	case s.queue <- j:
		return true
	default:
		s.dropped.Add(1)
		return false
	}
}

func (s *PayloadLogService) worker() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.cfg.FlushInterval)
	defer ticker.Stop()
	batch := make([]*payloadlog.PayloadLog, 0, s.cfg.BatchSize)
	var size int64
	flush := func() {
		if len(batch) == 0 {
			return
		}
		s.write(batch)
		batch = make([]*payloadlog.PayloadLog, 0, s.cfg.BatchSize)
		size = 0
	}
	for {
		select {
		case j, ok := <-s.queue:
			if !ok {
				flush()
				return
			}
			l := normalizePayloadLogJob(j)
			batch = append(batch, l)
			size += int64(len(l.RequestBody) + len(l.ResponseBody))
			if len(batch) >= s.cfg.BatchSize || size >= s.cfg.BatchMaxBytes {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}
func normalizePayloadLogJob(j PayloadLogJob) *payloadlog.PayloadLog {
	l := j.Log
	if l == nil {
		l = &payloadlog.PayloadLog{}
	}
	defer func() {
		if recover() != nil {
			l.RequestBody = `{"messages":[]}`
			l.ResponseBody = `{"messages":[]}`
		}
	}()
	p := payloadmsg.DetectProtocol(l.Path)
	rb, _ := payloadmsg.NormalizeRequest(p, l.RequestContentType, j.RequestRaw)
	sb, _ := payloadmsg.NormalizeResponse(p, l.ResponseContentType, j.ResponseRaw)
	l.RequestBody = string(rb)
	l.ResponseBody = string(sb)
	if l.ModelName == "" {
		var x struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(j.RequestRaw, &x)
		l.ModelName = x.Model
	}
	return l
}
func (s *PayloadLogService) write(batch []*payloadlog.PayloadLog) {
	if s.store == nil {
		s.dropped.Add(uint64(len(batch)))
		return
	}
	var err error
	for attempt := 0; attempt <= s.cfg.WriteRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.WriteTimeout)
		err = s.store.InsertBatch(ctx, batch)
		cancel()
		if err == nil {
			s.available.Store(true)
			s.storeErr.Store("")
			return
		}
		if attempt < s.cfg.WriteRetries {
			time.Sleep(time.Second << attempt)
		}
	}
	s.available.Store(false)
	s.storeErr.Store(err.Error())
	s.dropped.Add(uint64(len(batch)))
	slog.Error("payload log batch dropped", "count", len(batch), "error", err)
}
func (s *PayloadLogService) List(ctx context.Context, f payloadlog.PayloadLogFilter, p payloadlog.Pagination) ([]payloadlog.PayloadLogMeta, int64, error) {
	if !s.available.Load() || s.store == nil {
		return nil, 0, ErrPayloadLogStoreUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.QueryTimeout)
	defer cancel()
	return s.store.List(ctx, f, p)
}
func (s *PayloadLogService) GetByRequestID(ctx context.Context, id string, tr *payloadlog.TimeRange) ([]*payloadlog.PayloadLog, error) {
	if !s.available.Load() || s.store == nil {
		return nil, ErrPayloadLogStoreUnavailable
	}
	id = strings.TrimPrefix(id, "client:")
	ctx, cancel := context.WithTimeout(ctx, s.cfg.QueryTimeout)
	defer cancel()
	return s.store.GetByRequestID(ctx, id, tr)
}
func (s *PayloadLogService) Close() {
	if s == nil || !s.cfg.Enabled || !s.stopped.CompareAndSwap(false, true) {
		return
	}
	close(s.queue)
	s.wg.Wait()
	if s.store != nil {
		_ = s.store.Close()
	}
}
