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

package pipeline

import (
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
)

// Config is the configuration of a pipeline.
type Config []operator.Config

// BuildOperators builds the operators from the list of configs into operators.
func (c Config) BuildOperators(bc operator.BuildContext) ([]operator.Operator, error) {
	// buildsMulti's key represents an operator's ID that builds multiple operators, e.g. Plugins.
	// the value is the plugin's first operator's ID.
	buildsMulti := make(map[string]string)
	operators := make([]operator.Operator, 0, len(c))
	for _, builder := range c {
		op, err := builder.Build(bc)
		if err != nil {
			return nil, err
		}

		if builder.BuildsMultipleOps() {
			buildsMulti[bc.PrependNamespace(builder.ID())] = op[0].ID()
		}
		operators = append(operators, op...)
	}
	if err := SetOutputIDs(operators, buildsMulti); err != nil {
		return nil, err
	}

	return operators, nil
}

// BuildPipeline will build a pipeline from the config.
func (c Config) BuildPipeline(bc operator.BuildContext, defaultOperator operator.Operator) (*DirectedPipeline, error) {
	if defaultOperator != nil {
		bc.DefaultOutputIDs = []string{defaultOperator.ID()}
	}

	operators, err := c.BuildOperators(bc)
	if err != nil {
		return nil, err
	}

	if defaultOperator != nil {
		operators = append(operators, defaultOperator)
	}

	return NewDirectedPipeline(operators)
}

// SetOutputIDs Loops through all the operators and sets a default output to the next operator in the slice.
// Additionally, if the output is set to a plugin, it sets the output to the first operator in the plugins pipeline.
func SetOutputIDs(operators []operator.Operator, buildsMulti map[string]string) error {
	for i, op := range operators {
		// because no output is specified at this point for the last operator,
		// it will always be empty and there is nothing after it to automatically point towards, so we break the loop
		if i+1 == len(operators) {
			break
		}

		if len(op.GetOutputIDs()) == 0 {
			op.SetOutputIDs([]string{operators[i+1].ID()})
			continue
		}

		// Check if there are any plugins within the outputIDs of the operator. If there are, change the output to be the first op in the plugin
		allOutputs := []string{}
		pluginFound := false
		for _, id := range op.GetOutputIDs() {
			if pid, ok := buildsMulti[id]; ok {
				id = pid
				pluginFound = true
			}
			allOutputs = append(allOutputs, id)
		}

		if pluginFound {
			op.SetOutputIDs(allOutputs)
		}
	}
	return nil
}
