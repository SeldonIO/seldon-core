/*
Copyright 2022 Seldon Technologies Ltd.

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

package experiment

import "fmt"

type ExperimentNotFound struct {
	experimentName string
}

func (enf *ExperimentNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ExperimentNotFound)
	return ok
}

func (enf *ExperimentNotFound) Error() string {
	return fmt.Sprintf("Experiment not found %s", enf.experimentName)
}

type ExperimentBaselineExists struct {
	experimentName string
	name           string
}

func (ebe *ExperimentBaselineExists) Error() string {
	return fmt.Sprintf("Resource %s already in experiment %s as a baseline. A model or pipeline can only appear in one experiment as a baseline", ebe.name, ebe.experimentName)
}

type ExperimentNoCandidatesOrMirrors struct {
	experimentName string
}

func (enc *ExperimentNoCandidatesOrMirrors) Error() string {
	return fmt.Sprintf("experiment %s has no candidates or mirror", enc.experimentName)
}

type ExperimentDefaultNotFound struct {
	experimentName  string
	defaultResource string
}

func (enc *ExperimentDefaultNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ExperimentDefaultNotFound)
	return ok
}

func (enc *ExperimentDefaultNotFound) Error() string {
	return fmt.Sprintf("default model/pipeline %s not found in experiment %s candidates", enc.defaultResource, enc.experimentName)
}

type ExperimentNoDuplicates struct {
	experimentName string
	resource       string
}

func (enc *ExperimentNoDuplicates) Is(tgt error) bool {
	_, ok := tgt.(*ExperimentNoDuplicates)
	return ok
}

func (enc *ExperimentNoDuplicates) Error() string {
	return fmt.Sprintf("each candidate and mirror must be unique but found resource %s duplicated in experiment %s", enc.resource, enc.experimentName)
}
