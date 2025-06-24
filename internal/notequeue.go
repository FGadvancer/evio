// Copyright 2018 Joshua J Baker. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package internal

import "sync"

// this is a good candiate for a lock-free structure.

//⚙️ 延迟触发消费（第一次 Add 时才调度）
//
//👥 多个 goroutine 并发写入，但只希望一次触发调度
//
//📦 批处理 note 提高吞吐量
//
//♻️ 尽量重用内存减少 GC 压力
//这就是为什么这里不用channel+for select的方式

type noteQueue struct {
	mu    sync.Mutex
	notes []interface{}
}

func (q *noteQueue) Add(note interface{}) (one bool) {
	q.mu.Lock()
	q.notes = append(q.notes, note)
	n := len(q.notes)
	q.mu.Unlock()
	return n == 1
}

func (q *noteQueue) ForEach(iter func(note interface{}) error) error {
	q.mu.Lock()
	if len(q.notes) == 0 {
		q.mu.Unlock()
		return nil
	}
	notes := q.notes
	q.notes = nil
	q.mu.Unlock()
	for _, note := range notes {
		if err := iter(note); err != nil {
			return err
		}
	}
	q.mu.Lock()
	if q.notes == nil {
		for i := range notes {
			notes[i] = nil
		}
		q.notes = notes[:0]
	}
	q.mu.Unlock()
	return nil
}
