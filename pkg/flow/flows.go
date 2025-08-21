// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Package flow runs automated workflows: trigger → tasks (with AI parsers) → notifications.
//
// A flow needs a trigger to start. In the middle you can run multiple tasks,
// with AI parsers in between to interpret or transform data. At the end you
// can plug in one or more notification integrations (e.g. Slack, email).
package flow
