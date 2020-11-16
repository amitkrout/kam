/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"fmt"
	"strings"

	"github.com/tektoncd/pipeline/pkg/substitution"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

var validWhenOperators = []string{
	string(selection.In),
	string(selection.NotIn),
}

func (wes WhenExpressions) validate() *apis.FieldError {
	if err := wes.validateWhenExpressionsFields(); err != nil {
		return err
	}
	if err := wes.validateTaskResultsVariables(); err != nil {
		return err
	}
	return nil
}

func (wes WhenExpressions) validateWhenExpressionsFields() *apis.FieldError {
	for _, we := range wes {
		if err := we.validateWhenExpressionFields(); err != nil {
			return err
		}
	}
	return nil
}

func (we *WhenExpression) validateWhenExpressionFields() *apis.FieldError {
	if equality.Semantic.DeepEqual(we, &WhenExpression{}) || we == nil {
		return apis.ErrMissingField(apis.CurrentField)
	}
	if !sets.NewString(validWhenOperators...).Has(string(we.Operator)) {
		message := fmt.Sprintf("operator %q is not recognized. valid operators: %s", we.Operator, strings.Join(validWhenOperators, ","))
		return apis.ErrInvalidValue(message, "spec.task.when")
	}
	if len(we.Values) == 0 {
		return apis.ErrInvalidValue("expecting non-empty values field", "spec.task.when")
	}
	return nil
}

func (wes WhenExpressions) validateTaskResultsVariables() *apis.FieldError {
	for _, we := range wes {
		expressions, ok := we.GetVarSubstitutionExpressions()
		if ok {
			if LooksLikeContainsResultRefs(expressions) {
				expressions = filter(expressions, looksLikeResultRef)
				resultRefs := NewResultRefs(expressions)
				if len(expressions) != len(resultRefs) {
					message := fmt.Sprintf("expected all of the expressions %v to be result expressions but only %v were", expressions, resultRefs)
					return apis.ErrInvalidValue(message, "spec.tasks.when")
				}
			}
		}
	}
	return nil
}

func (wes WhenExpressions) validatePipelineParametersVariables(prefix string, paramNames sets.String, arrayParamNames sets.String) *apis.FieldError {
	for _, we := range wes {
		if err := validateStringVariable(fmt.Sprintf("input[%s]", we.Input), we.Input, prefix, paramNames, arrayParamNames); err != nil {
			return err
		}
		for _, val := range we.Values {
			if err := validateStringVariable(fmt.Sprintf("values[%s]", val), val, prefix, paramNames, arrayParamNames); err != nil {
				return err
			}
		}
	}
	return nil
}
func validateStringVariable(name, value, prefix string, stringVars sets.String, arrayVars sets.String) *apis.FieldError {
	if err := substitution.ValidateVariable(name, value, prefix, "task when expression", "pipelinespec.when", stringVars); err != nil {
		return err
	}
	if err := substitution.ValidateVariableProhibited(name, value, prefix, "task when expression", "pipelinespec.when", arrayVars); err != nil {
		return err
	}
	return nil
}
