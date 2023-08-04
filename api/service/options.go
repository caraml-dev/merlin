// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import "gorm.io/gorm"

type ListOptions interface {
	apply(q *gorm.DB) *gorm.DB
}

func LimitOptions(limit int) ListOptions {
	return &limitOptions{limit: limit}
}

type limitOptions struct {
	limit int
}

func (l *limitOptions) apply(q *gorm.DB) *gorm.DB {
	return q.Limit(l.limit)
}

func OffsetOptions(offset int) ListOptions {
	return &offsetOptions{offset: offset}
}

type offsetOptions struct {
	offset int
}

func (o *offsetOptions) apply(q *gorm.DB) *gorm.DB {
	return q.Offset(o.offset)
}
