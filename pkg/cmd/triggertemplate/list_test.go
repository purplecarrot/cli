// Copyright © 2019 The Tekton Authors.
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

package triggertemplate

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/test"
	cb "github.com/tektoncd/cli/pkg/test/builder"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggertest "github.com/tektoncd/triggers/test"
	tb "github.com/tektoncd/triggers/test/builder"
	"gotest.tools/v3/golden"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListTriggerTemplate(t *testing.T) {
	now := time.Now()

	ns := []*corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "random",
			},
		},
	}

	tts := []*v1alpha1.TriggerTemplate{
		tb.TriggerTemplate("tt1", "foo", cb.TriggerTemplateCreationTime(now.Add(-2*time.Minute))),
		tb.TriggerTemplate("tt2", "foo", cb.TriggerTemplateCreationTime(now.Add(-30*time.Second))),
		tb.TriggerTemplate("tt3", "foo", cb.TriggerTemplateCreationTime(now.Add(-200*time.Hour))),
		tb.TriggerTemplate("tt4", "foo"),
	}

	tests := []struct {
		name      string
		command   *cobra.Command
		args      []string
		wantError bool
	}{
		{
			name:      "Invalid namespace",
			command:   command(t, tts, now, ns),
			args:      []string{"list", "-n", "default"},
			wantError: true,
		},
		{
			name:      "No TriggerTemplate",
			command:   command(t, tts, now, ns),
			args:      []string{"list", "-n", "random"},
			wantError: false,
		},
		{
			name:      "Multiple TriggerTemplates",
			command:   command(t, tts, now, ns),
			args:      []string{"list", "-n", "foo"},
			wantError: false,
		},
		{
			name:      "Multiple TriggerTemplates with output format",
			command:   command(t, tts, now, ns),
			args:      []string{"list", "-n", "foo", "-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}"},
			wantError: false,
		},
	}

	for _, td := range tests {
		t.Run(td.name, func(t *testing.T) {
			got, err := test.ExecuteCommand(td.command, td.args...)

			if err != nil && !td.wantError {
				t.Errorf("Unexpected error: %v", err)
			}
			golden.Assert(t, got, strings.ReplaceAll(fmt.Sprintf("%s.golden", t.Name()), "/", "-"))
		})
	}
}

func command(t *testing.T, tts []*v1alpha1.TriggerTemplate, now time.Time, ns []*corev1.Namespace) *cobra.Command {
	// fake clock advanced by 1 hour
	clock := clockwork.NewFakeClockAt(now)

	cs := test.SeedTestResources(t, triggertest.Resources{TriggerTemplates: tts, Namespaces: ns})

	p := &test.Params{Tekton: cs.Pipeline, Clock: clock, Kube: cs.Kube, Triggers: cs.Triggers}

	return Command(p)
}
