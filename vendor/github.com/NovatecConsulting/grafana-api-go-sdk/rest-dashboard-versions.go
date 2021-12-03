package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

//Version struct that wraps GetAllDashboardVersions API Call Response
type Version struct {
	Id            int       `json:"id"`
	DashboardId   int       `json:"dashboardId"`
	ParentVersion int       `json:"parentVersion"`
	RestoredFrom  int       `json:"restoredFrom"`
	Version       int       `json:"version"`
	Created       time.Time `json:"created"`
	CreatedBy     string    `json:"createdBy"`
	Message       string    `json:"message"`
}

// GetAllDashboardVersions loads the first "limit" dashboard versions of a specific dashboard by id.
//
// Reflects GET /api/dashboards/id/:dashboardId/versions API call.
func (r *Client) GetAllDashboardVersions(ctx context.Context, id int, limit int) ([]Version, error) {
	var (
		raw       []byte
		versions  []Version
		code      int
		err       error
		requestParams = make(url.Values)
	)
	requestParams.Add("limit", strconv.Itoa(limit))
	if raw, code, err = r.get(ctx, fmt.Sprintf("/api/dashboards/id/%d/versions", id), requestParams); err != nil {
		return nil, err
	}
	if code != 200 {
		return versions, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&versions); err != nil {
		return nil, err
	}
	//err = json.Unmarshal(raw, &versions)
	return versions, err
}

