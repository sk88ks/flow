package flow

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	asserts := assert.New(t)

	numCPU := runtime.NumCPU()

	tests := []struct {
		name string
		want *Flow
	}{
		{
			name: "Create A New Flow",
			want: &Flow{
				concurrencyLevel: numCPU,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New()
			asserts.Equal(tt.want, got)
		})
	}
}

func TestFlow_SetConcurrencyLevel(t *testing.T) {
	asserts := assert.New(t)

	type args struct {
		l int
	}
	tests := []struct {
		name                 string
		args                 args
		wantCuncurrencyLevel int
	}{
		{
			name: "Set Concurrency Level",
			args: args{
				l: 2,
			},
			wantCuncurrencyLevel: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flow{}
			f.SetConcurrencyLevel(tt.args.l)
			asserts.Equal(tt.wantCuncurrencyLevel, f.concurrencyLevel)
		})
	}
}

type TestProcess struct {
	Count  int
	isWait bool
	mu     *sync.RWMutex
}

func (p *TestProcess) CountUp() error {
	if p.isWait {
		time.Sleep(2 * time.Second)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Count++
	if p.Count > 2 {
		p.Count = -1
		return errors.New("Overflow")
	}
	return nil
}

func TestFlow_Serial(t *testing.T) {
	asserts := assert.New(t)

	type fields struct {
		concurrencyLevel int
	}
	type args struct {
		ctx   context.Context
		fsNum int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
		wantErr   error
	}{
		{
			name: "All Funcs End Well",
			fields: fields{
				concurrencyLevel: 1,
			},
			args: args{
				ctx:   context.Background(),
				fsNum: 2,
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name: "A Func Issues Error",
			fields: fields{
				concurrencyLevel: 1,
			},
			args: args{
				ctx:   context.Background(),
				fsNum: 3,
			},
			wantCount: -1,
			wantErr:   errors.New("Overflow"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flow{
				concurrencyLevel: tt.fields.concurrencyLevel,
			}
			p := &TestProcess{
				mu: &sync.RWMutex{},
			}
			fs := []Func{nil}
			for i := 0; i < tt.args.fsNum; i++ {
				fs = append(fs, WrapFunc(p.CountUp))
			}
			err := f.Serial(fs...)(tt.args.ctx)
			asserts.Equal(tt.wantErr, err)
			asserts.Equal(tt.wantCount, p.Count)
		})
	}
}

func TestFlow_Parallel(t *testing.T) {
	asserts := assert.New(t)

	ctxWithTimeout, _ := context.WithTimeout(context.Background(), 1*time.Second)

	type fields struct {
		concurrencyLevel int
		mu               *sync.RWMutex
		isWait           bool
	}
	type args struct {
		ctx   context.Context
		fsNum int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
		wantErr   error
	}{
		{
			name: "All Funcs End Well",
			fields: fields{
				concurrencyLevel: 3,
				mu:               &sync.RWMutex{},
			},
			args: args{
				ctx:   context.Background(),
				fsNum: 2,
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name: "A Func Issues Error",
			fields: fields{
				concurrencyLevel: 3,
				mu:               &sync.RWMutex{},
			},
			args: args{
				ctx:   context.Background(),
				fsNum: 3,
			},
			wantCount: -1,
			wantErr:   errors.New("Overflow"),
		},
		{
			name: "Call Ctx Done",
			fields: fields{
				concurrencyLevel: 3,
				mu:               &sync.RWMutex{},
				isWait:           true,
			},
			args: args{
				ctx:   ctxWithTimeout,
				fsNum: 2,
			},
			wantCount: 0,
			wantErr:   context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Flow{
				concurrencyLevel: tt.fields.concurrencyLevel,
			}
			p := &TestProcess{
				mu:     tt.fields.mu,
				isWait: tt.fields.isWait,
			}
			fs := []Func{nil}
			for i := 0; i < tt.args.fsNum; i++ {
				fs = append(fs, WrapFunc(p.CountUp))
			}
			err := f.Parallel(fs...)(tt.args.ctx)
			asserts.Equal(tt.wantErr, err)
			asserts.Equal(tt.wantCount, p.Count)
		})
	}
}
