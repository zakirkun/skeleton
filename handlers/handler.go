package handlers

import (
	"context"

	events "github.com/crowdeco/skeleton/events"
	paginations "github.com/crowdeco/skeleton/paginations"
	adapter "github.com/crowdeco/skeleton/paginations/adapter"
	services "github.com/crowdeco/skeleton/services"
	elastic "github.com/olivere/elastic/v7"
)

const PAGINATION_EVENT = "event.pagination"
const BEFORE_CREATE_EVENT = "event.before_create"
const AFTER_CREATE_EVENT = "event.after_create"
const BEFORE_UPDATE_EVENT = "event.before_update"
const AFTER_UPDATE_EVENT = "event.after_update"
const BEFORE_DELETE_EVENT = "event.before_delete"
const AFTER_DELETE_EVENT = "event.after_delete"

type Handler struct {
	Context       context.Context
	Elasticsearch *elastic.Client
	Dispatcher    *events.Dispatcher
	Repository    *services.Repository
}

func (h *Handler) Paginate(paginator paginations.Pagination) (paginations.PaginationMeta, []interface{}) {
	query := elastic.NewBoolQuery()

	h.Dispatcher.Dispatch(PAGINATION_EVENT, &events.Pagination{
		Repository: h.Repository,
		Query:      query,
		Filters:    paginator.Filters,
	})

	var result []interface{}
	adapter := adapter.NewElasticsearchAdapter(h.Context, h.Elasticsearch, paginator.Model, query)
	paginator.Paginate(adapter)
	paginator.Pager.Results(&result)
	next := paginator.Page + 1
	total, _ := paginator.Pager.Nums()

	if paginator.Page*paginator.Limit > int(total) {
		next = -1
	}

	return paginations.PaginationMeta{
		Record:   len(result),
		Page:     paginator.Page,
		Previous: paginator.Page - 1,
		Next:     next,
		Limit:    paginator.Limit,
		Total:    int(total),
	}, result
}

func (h *Handler) Create(v interface{}) error {
	h.Repository.StartTransaction()
	h.Dispatcher.Dispatch(BEFORE_CREATE_EVENT, &events.Model{
		Data:       v,
		Repository: h.Repository,
	})

	err := h.Repository.Create(v)
	if err != nil {
		h.Repository.Rollback()

		return err
	}

	h.Dispatcher.Dispatch(AFTER_CREATE_EVENT, &events.Model{
		Data:       v,
		Repository: h.Repository,
	})
	h.Repository.Commit()

	return nil
}

func (h *Handler) Update(v interface{}, id string) error {
	h.Repository.StartTransaction()
	h.Dispatcher.Dispatch(BEFORE_UPDATE_EVENT, &events.Model{
		Id:         id,
		Data:       v,
		Repository: h.Repository,
	})

	err := h.Repository.Update(v)
	if err != nil {
		h.Repository.Rollback()

		return err
	}

	h.Dispatcher.Dispatch(AFTER_UPDATE_EVENT, &events.Model{
		Id:         id,
		Data:       v,
		Repository: h.Repository,
	})
	h.Repository.Commit()

	return nil
}

func (h *Handler) Bind(v interface{}, id string) error {
	return h.Repository.Bind(v, id)
}

func (h *Handler) All(v interface{}) error {
	return h.Repository.All(v)
}

func (h *Handler) Delete(v interface{}, id string) error {
	h.Repository.StartTransaction()
	h.Dispatcher.Dispatch(BEFORE_DELETE_EVENT, &events.Model{
		Id:         id,
		Data:       v,
		Repository: h.Repository,
	})

	err := h.Repository.Delete(v, id)
	if err != nil {
		h.Repository.Rollback()

		return err
	}

	h.Dispatcher.Dispatch(AFTER_DELETE_EVENT, &events.Model{
		Id:         id,
		Data:       v,
		Repository: h.Repository,
	})
	h.Repository.Commit()

	return nil
}
