package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/letsblockit/letsblockit/src/db"
	"github.com/letsblockit/letsblockit/src/filters"
	"github.com/letsblockit/letsblockit/src/users/auth"
	"gopkg.in/yaml.v3"
)

const listExportTemplate = `# letsblock.it filter list export
#
# List token: %s
# Export date: %s
#
# You can edit this file and render it locally, check out instructions at:
# https://github.com/letsblockit/letsblockit/tree/main/cmd/render/README.md

`

const renderListSuffix = ".txt"
const installPromptFilterTemplate = `
! Hide the list install prompt for that list
%s###install-prompt-%s
`

func (s *Server) renderList(c echo.Context) error {
	token, err := uuid.Parse(strings.TrimSuffix(c.Param("token"), renderListSuffix))
	if err != nil {
		return echo.ErrNotFound
	}

	// In order to reduce resource consumption, we compute an etag based on:
	//   - a hash of the filter templates
	//   - the latest change to any parameter in the list
	requestETag, listETag := getEtag(c), s.filterHash
	etagPresent, etagMatch := requestETag != "", false

	var storedList db.GetListForTokenRow
	var storedInstances []db.GetInstancesForListRow
	if err := s.store.RunTx(c, func(ctx context.Context, q db.Querier) error {
		var e error
		storedList, e = q.GetListForToken(ctx, token)
		switch {
		case e == db.NotFound:
			return echo.ErrNotFound
		case e != nil:
			return fmt.Errorf("failed to get list: %w", e)
		case s.bans.IsBanned(storedList.UserID):
			return echo.ErrForbidden
		}

		if c.Request().Header.Get("Referer") == "" {
			e = q.MarkListDownloaded(ctx, token)
			if e != nil {
				return fmt.Errorf("failed to mark list download: %w", e)
			}
		}

		if ts, ok := storedList.LastUpdated.(time.Time); ok {
			listETag += ts.UTC().Format("15040520060102")
		}
		etagMatch = listETag == requestETag
		if etagMatch {
			return nil
		}

		storedInstances, e = q.GetInstancesForList(ctx, storedList.ID)
		if e != nil {
			return fmt.Errorf("failed to get instances: %w", e)
		}
		return nil
	}); err != nil {
		return err
	}

	_ = s.statsd.Incr("letsblockit.list_download", []string{
		fmt.Sprintf("etag_present:%t", etagPresent),
		fmt.Sprintf("etag_match:%t", etagMatch),
	}, 1)
	if etagMatch {
		return c.NoContent(http.StatusNotModified)
	}

	c.Response().Header().Set("Etag", listETag)

	list, err := convertFilterList(storedInstances)
	if err != nil {
		return fmt.Errorf("failed to convert list: %w", err)
	}
	if _, ok := c.QueryParams()["test_mode"]; ok {
		list.TestMode = true
	}

	if err = list.Render(c.Response(), c.Logger(), s.filters); err != nil {
		return fmt.Errorf("failed to render list: %w", err)
	}

	if s.options.OfficialInstance {
		_, err = fmt.Fprintf(c.Response(), installPromptFilterTemplate, mainDomain, token)
	} else {
		_, err = fmt.Fprintf(c.Response(), installPromptFilterTemplate, c.Request().Host, token)
	}

	return err
}

func (s *Server) exportList(c echo.Context) error {
	token, err := uuid.Parse(c.Param("token"))
	if err != nil {
		return echo.ErrNotFound
	}

	var storedInstances []db.GetInstancesForListRow
	if err := s.store.RunTx(c, func(ctx context.Context, q db.Querier) error {
		storedList, e := q.GetListForToken(ctx, token)
		switch {
		case e == db.NotFound:
			return echo.ErrNotFound
		case e != nil:
			return e
		case auth.GetUserId(c) != storedList.UserID:
			return echo.ErrForbidden
		}
		storedInstances, e = q.GetInstancesForList(ctx, storedList.ID)
		return e
	}); err != nil {
		return err
	}

	list, err := convertFilterList(storedInstances)
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Type", "text/yaml")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=\"exported-filter-list.yaml\"")
	c.Response().WriteHeader(200)
	_, err = fmt.Fprintf(c.Response(), listExportTemplate, token, s.now().Format("2006-01-02"))
	if err != nil {
		return nil
	}
	err = yaml.NewEncoder(c.Response()).Encode(&list)
	if err != nil {
		return nil
	}
	return nil
}

func convertFilterList(storedInstances []db.GetInstancesForListRow) (*filters.List, error) {
	list := &filters.List{Title: "My filters"}
	var customFilterInstances []*filters.Instance
	for _, storedInstance := range storedInstances {
		instance := &filters.Instance{
			Template: storedInstance.TemplateName,
			Params:   make(map[string]interface{}),
			TestMode: storedInstance.TestMode,
		}
		err := storedInstance.Params.AssignTo(&instance.Params)
		if err != nil {
			return nil, err
		}
		if instance.Template == filters.CustomRulesFilterName {
			customFilterInstances = append(customFilterInstances, instance)
		} else {
			list.Instances = append(list.Instances, instance)
		}
	}
	if len(customFilterInstances) > 0 {
		list.Instances = append(list.Instances, customFilterInstances...)
	}
	return list, nil
}
