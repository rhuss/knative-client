// Copyright © 2019 The Knative Authors
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

package binding

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	v1alpha12 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewBindingDescribeCommand returns a new command for describe a sink binding object
func NewBindingDescribeCommand(p *commands.KnParams) *cobra.Command {

	bindingDescribeCommand := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a sink binding.",
		Example: `
  # Describe a sink binding with name 'mysinkbinding'
  kn source binding describe mysinkbinding`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source binding describe' requires the name of the source as argument")
			}
			name := args[0]

			bindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			binding, err := bindingClient.GetSinkBinding(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeSinkBinding(dw, binding, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, binding.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := bindingDescribeCommand.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")

	return bindingDescribeCommand
}

func writeSinkBinding(dw printers.PrefixWriter, binding *v1alpha12.SinkBinding, printDetails bool) {
	commands.WriteMetadata(dw, &binding.ObjectMeta, printDetails)
	writeSubject(dw, binding)
	writeSink(dw, &binding.Spec.Sink)
}

func writeSink(dw printers.PrefixWriter, sink *duckv1.Destination) {
	subWriter := dw.WriteAttribute("Sink", "")
	subWriter.WriteAttribute("Name", sink.Ref.Name)
	subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}

func writeSubject(dw printers.PrefixWriter, binding *v1alpha12.SinkBinding) {
	subject := binding.Spec.Subject
	subjectDw := dw.WriteAttribute("Subject", "")
	subjectDw.WriteAttribute("Kind", subject.Kind)
	subjectDw.WriteAttribute("APIVersion", subject.APIVersion)
	if subject.Name != "" {
		subjectDw.WriteAttribute("Name", subject.Name)
	}
	if subject.Selector != nil {
		matchDw := subjectDw.WriteAttribute("Selector", "")
		selector := subject.Selector
		if len(selector.MatchLabels) > 0 {
			var lKeys []string
			for k := range selector.MatchLabels {
				lKeys = append(lKeys, k)
			}
			sort.Strings(lKeys)
			for _, k := range lKeys {
				matchDw.WriteAttribute(k, selector.MatchLabels[k])
			}
		}
		// TOOD: Print out selector.MatchExpressions
	}
}