package k8s

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetNestedSliceStringInterfaceMap returns a slice of string to interface maps within the given unstructured
// at the path described by the given variadic arguments. This is analogous to the suite of "NestedThing"
// functions in apimachinery.
func GetNestedSliceStringInterfaceMap(
	u *unstructured.Unstructured, path ...string,
) ([]map[string]interface{}, bool, error) {
	strPath := strings.Join(path, "/")
	untyped, found, err := unstructured.NestedSlice(u.Object, path...)
	if !found {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, fmt.Errorf("expected slice for %s, got something different: %s", strPath, err)
	}
	typed := []map[string]interface{}{}
	for _, utv := range untyped {
		tv, ok := utv.(map[string]interface{})
		if !ok {
			return nil, true, fmt.Errorf("value in slice %s is not string interface map: %s", strPath, utv)
		}
		typed = append(typed, tv)
	}

	return typed, true, nil
}

func pullStringValue(m map[string]interface{}, key string) (string, bool, error) {
	if m == nil {
		return "", false, fmt.Errorf("map is nil while reading %s", key)
	}
	value, ok := m[key]
	if !ok {
		return "", false, nil
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", true, fmt.Errorf("value for %s is not a string: %#v", key, value)
	}
	return stringValue, true, nil
}

// DoesContainerStatusShowCompleted takes the given containerStatus map and returns whether or not the status shows
// the container as having terminated with the reason "Completed".
func DoesContainerStatusShowCompleted(containerStatus map[string]interface{}) (bool, error) {
	reason, reasonFound, err := unstructured.NestedString(
		containerStatus, "state", "terminated", "reason",
	)
	if err != nil {
		return false, err
	}
	if reasonFound && reason == "Completed" {
		return true, nil
	}
	return false, nil
}

// DoesConditionShowReady returns if the given map reflects a condition that is in a ready state.
// For instance, if the condition has "ContainersReady" as true but "PIDPressure" as true as well, the
// condition is showing not ready.
// Essentially just a big switch statement.
func DoesConditionShowReady(condition map[string]interface{}) (bool, error) {
	ct, ctFound, err := pullStringValue(condition, "type")
	if err != nil {
		return false, err
	}
	if !ctFound {
		return false, fmt.Errorf("condition missing type: %#v", condition)
	}
	cs, csFound, err := pullStringValue(condition, "status")
	if err != nil {
		return false, err
	}
	if !csFound {
		return false, fmt.Errorf("condition missing status: %#v", condition)
	}
	switch ct {
	case "Ready", "Initialized", "ContainersReady", "PodScheduled", "Progressing", "Available", "PodReadyToStartContainers":
		return cs == string(v1.ConditionTrue), nil
	case "MemoryPressure", "DiskPressure", "PIDPressure", "NetworkUnavailable", "FrequentDockerRestart",
		"FrequentContainerdRestart", "KernelDeadlock", "ReadonlyFilesystem", "CorruptDockerOverlay2",
		"FrequentUnregisterNetDevice", "FrequentKubeletRestart":
		return cs == string(v1.ConditionFalse), nil
	default:
		return false, fmt.Errorf("unknown condition: %s", ct)
	}
}

// DoesPhaseShowReady is similar to DoesConditionShowReady but works for phases.
// Requires the kind of resource the phase is attached too, as different resource have different phase strings.
// Currently supports Pods, Namespaces and PVCs.
func DoesPhaseShowReady(phase string, kind string) (bool, error) {
	switch kind {
	case "Pod":
		return phase == "Running", nil
	case "Namespace":
		return phase == "Active", nil
	case "PersistentVolumeClaim":
		return phase == "Pending" || phase == "Bound", nil
	}
	return false, fmt.Errorf("kind is either unknown or does not have a phase: %s", kind)
}

// CheckUnstructuredForReadyState takes in an *unstructured.Unstructured and returns a boolean describe if or if not
// the resource should be considered ready, depending on its (if applicable): conditions, phase, containers.
// It also logs the result as a pretty string.
func CheckUnstructuredForReadyState(logger *log.Logger, resource *unstructured.Unstructured) (bool, error) {
	// Gather information
	name := resource.GetName()
	kind := resource.GetKind()
	namespace := resource.GetNamespace()
	phase, phaseFound, err := unstructured.NestedString(resource.Object, "status", "phase")
	if err != nil {
		return false, err
	}
	conditions, conditionsFound, err := GetNestedSliceStringInterfaceMap(resource, "status", "conditions")
	if err != nil {
		return false, err
	}
	containerStatuses, containerStatusesFound, err := GetNestedSliceStringInterfaceMap(
		resource, "status", "containerStatuses",
	)
	if err != nil {
		return false, err
	}

	// Build info string
	statusStringBuilder := strings.Builder{}
	statusStringBuilder.WriteString(kind + ": " + name + ": ")
	if namespace != "" {
		statusStringBuilder.WriteString("(" + namespace + "): ")
	}

	// Now we do our checks
	resourceIsReady := true
	addComma := false

	if phaseFound && kind != "" {
		addComma = true
		logger.WithFields(log.Fields{
			"phase": phase,
			"kind":  kind,
		}).Debug("evaluating phase")
		statusStringBuilder.WriteString("Phase: " + phase)
		phaseShowsReady, err := DoesPhaseShowReady(phase, kind)
		if err != nil {
			return false, err
		}
		if !phaseShowsReady {
			resourceIsReady = false
		}
	}

	badConditions := []string{}
	if conditionsFound {
		if addComma {
			statusStringBuilder.WriteString(", ")
		}
		addComma = true
		statusStringBuilder.WriteString("Conditions: ")
		logger.WithField("conditions", conditions).Debug("evaluating conditions")
		for i, condition := range conditions {
			conditionType, _, err := pullStringValue(condition, "type")
			if err != nil {
				return false, err
			}
			conditionStatus, _, err := pullStringValue(condition, "status")
			if err != nil {
				return false, err
			}
			if conditionStatus == string(v1.ConditionTrue) {
				statusStringBuilder.WriteString("+")
			} else {
				statusStringBuilder.WriteString("-")
			}
			conditionShowsReady, err := DoesConditionShowReady(condition)
			if err != nil {
				return false, err
			}
			if !conditionShowsReady {
				badConditions = append(badConditions, conditionType)
				resourceIsReady = false
			}
			statusStringBuilder.WriteString(conditionType)
			if i < len(conditions)-1 {
				statusStringBuilder.WriteString(" ")
			}
		}
	}

	if containerStatusesFound {
		if addComma {
			statusStringBuilder.WriteString(", ")
		}
		addComma = true
		statusStringBuilder.WriteString("Containers: ")
		for _, containerStatus := range containerStatuses {
			name, _, err := pullStringValue(containerStatus, "name")
			if err != nil {
				return false, err
			}
			ready, readyFound, err := unstructured.NestedBool(containerStatus, "ready")
			if err != nil {
				return false, err
			}
			if !readyFound {
				return false, fmt.Errorf("container status missing ready flag: %#v", containerStatus)
			}
			if ready {
				statusStringBuilder.WriteString("+")
			} else {
				completed, err := DoesContainerStatusShowCompleted(containerStatus)
				if err != nil {
					return false, err
				}
				if completed {
					statusStringBuilder.WriteString("*")
					allBadConditionsAreCompletedReadySignals := true
					for _, badCondition := range badConditions {
						if badCondition != "Ready" && badCondition != "ContainersReady" {
							allBadConditionsAreCompletedReadySignals = false
							break
						}
					}
					if allBadConditionsAreCompletedReadySignals {
						resourceIsReady = true
					}
				} else {
					statusStringBuilder.WriteString("-")
					resourceIsReady = false
				}
			}
			statusStringBuilder.WriteString(name + " ")
		}
	}

	statusString := statusStringBuilder.String()
	if resourceIsReady {
		logger.Debug(statusString)
	} else {
		logger.Error(statusString)
	}

	return resourceIsReady, nil
}
