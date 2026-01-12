// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package github

import (
	"time"
)

// CheckRun is a check run returned by the GitHub API.
type CheckRun struct {
	ID          int       `json:"id"`
	NodeID      string    `json:"node_id"`
	HeadSha     string    `json:"head_sha"`
	ExternalID  string    `json:"external_id"`
	URL         string    `json:"url"`
	HTMLURL     string    `json:"html_url"`
	DetailsURL  string    `json:"details_url"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Output      struct {
		Title            string `json:"title"`
		Summary          string `json:"summary"`
		Text             string `json:"text"`
		AnnotationsCount int    `json:"annotations_count"`
		AnnotationsURL   string `json:"annotations_url"`
	} `json:"output"`
	Name       string `json:"name"`
	CheckSuite struct {
		ID int `json:"id"`
	} `json:"check_suite"`
	App struct {
		ID     int    `json:"id"`
		NodeID string `json:"node_id"`
		Owner  struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"owner"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		ExternalURL string    `json:"external_url"`
		HTMLURL     string    `json:"html_url"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"app"`
	PullRequests []interface{} `json:"pull_requests"`
}

// CheckRunPayload is the JSON body for creating or updating a check run.
type CheckRunPayload struct {
	Name        string           `json:"name,omitempty"`
	HeadSha     string           `json:"head_sha,omitempty"`
	DetailsURL  string           `json:"details_url,omitempty"`
	Status      string           `json:"status,omitempty"`
	ExternalID  string           `json:"external_id,omitempty"`
	Conclusion  string           `json:"conclusion,omitempty"`
	StartedAt   string           `json:"started_at,omitempty"`
	CompletedAt string           `json:"completed_at,omitempty"`
	Output      *CheckRunOutput  `json:"output,omitempty"`
	Actions     []CheckRunAction `json:"actions,omitempty"`
}

// CheckRunOutput is the output object for create/update check run requests.
type CheckRunOutput struct {
	Title       string                  `json:"title,omitempty"`
	Summary     string                  `json:"summary,omitempty"`
	Text        string                  `json:"text,omitempty"`
	Annotations []CheckRunOutAnnotation `json:"annotations,omitempty"`
	Images      []CheckRunImage         `json:"images,omitempty"`
}

// CheckRunOutAnnotation is an annotation embedded in check run output (request).
type CheckRunOutAnnotation struct {
	Path            string `json:"path,omitempty"`
	StartLine       int    `json:"start_line,omitempty"`
	EndLine         int    `json:"end_line,omitempty"`
	StartColumn     int    `json:"start_column,omitempty"`
	EndColumn       int    `json:"end_column,omitempty"`
	AnnotationLevel string `json:"annotation_level,omitempty"`
	Message         string `json:"message,omitempty"`
	Title           string `json:"title,omitempty"`
	RawDetails      string `json:"raw_details,omitempty"`
}

// CheckRunImage is an image in check run output (request).
type CheckRunImage struct {
	Alt      string `json:"alt,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

// CheckRunAction is an action on a check run (request).
type CheckRunAction struct {
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Identifier  string `json:"identifier,omitempty"`
}
