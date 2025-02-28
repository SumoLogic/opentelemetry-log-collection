// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	context "context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	zap "go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	entry "github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
)

// NewMockOperator will return a basic operator mock
func NewMockOperator(id string) *Operator {
	mockOutput := &Operator{}
	mockOutput.On("ID").Return(id)
	mockOutput.On("CanProcess").Return(true)
	mockOutput.On("CanOutput").Return(true)
	return mockOutput
}

// FakeOutput is an empty output used primarily for testing
type FakeOutput struct {
	Received chan *entry.Entry
	*zap.SugaredLogger
}

// NewFakeOutput creates a new fake output with default settings
func NewFakeOutput(t testing.TB) *FakeOutput {
	return &FakeOutput{
		Received:      make(chan *entry.Entry, 100),
		SugaredLogger: zaptest.NewLogger(t).Sugar(),
	}
}

// CanOutput always returns false for a fake output
func (f *FakeOutput) CanOutput() bool { return false }

// CanProcess always returns true for a fake output
func (f *FakeOutput) CanProcess() bool { return true }

// ID always returns `fake` as the ID of a fake output operator
func (f *FakeOutput) ID() string { return "$.fake" }

// Logger returns the logger of a fake output
func (f *FakeOutput) Logger() *zap.SugaredLogger { return f.SugaredLogger }

// Outputs always returns nil for a fake output
func (f *FakeOutput) Outputs() []operator.Operator { return nil }

// Outputs always returns nil for a fake output
func (f *FakeOutput) GetOutputIDs() []string { return nil }

// SetOutputs immediately returns nil for a fake output
func (f *FakeOutput) SetOutputs(outputs []operator.Operator) error { return nil }

// SetOutputIDs immediately returns nil for a fake output
func (f *FakeOutput) SetOutputIDs(s []string) {}

// Start immediately returns nil for a fake output
func (f *FakeOutput) Start(_ operator.Persister) error { return nil }

// Stop immediately returns nil for a fake output
func (f *FakeOutput) Stop() error { return nil }

// Type always return `fake_output` for a fake output
func (f *FakeOutput) Type() string { return "fake_output" }

// Process will place all incoming entries on the Received channel of a fake output
func (f *FakeOutput) Process(ctx context.Context, entry *entry.Entry) error {
	f.Received <- entry
	return nil
}

// ExpectBody expects that a body will be received by the fake operator within a second
// and that it is equal to the given body
func (f *FakeOutput) ExpectBody(t testing.TB, body interface{}) {
	select {
	case e := <-f.Received:
		require.Equal(t, body, e.Body)
	case <-time.After(time.Second):
		require.FailNow(t, "Timed out waiting for entry")
	}
}

// ExpectEntry expects that an entry will be received by the fake operator within a second
// and that it is equal to the given body
func (f *FakeOutput) ExpectEntry(t testing.TB, expected *entry.Entry) {
	select {
	case e := <-f.Received:
		require.Equal(t, expected, e)
	case <-time.After(time.Second):
		require.FailNow(t, "Timed out waiting for entry")
	}
}
