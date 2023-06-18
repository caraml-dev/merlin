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

import (
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
)

type order string

const (
	descOrder order = "desc"
	ascOrder  order = "asc"
)

var orderMapping = map[order]paginator.Order{
	descOrder: paginator.DESC,
	ascOrder:  paginator.ASC,
}

func generatePagination(query PaginationQuery, cursorKeys []string, orderStrategy order) *paginator.Paginator {
	paginateEngine := paginator.New()
	paginateEngine.SetKeys(cursorKeys...)
	paginateEngine.SetLimit(query.Limit)
	paginateEngine.SetAfterCursor(query.Cursor)
	if strategy, ok := orderMapping[orderStrategy]; ok {
		paginateEngine.SetOrder(strategy)
	}
	return paginateEngine
}
