package health

func Aggregate(results []Result) Status {
	if len(results) == 0 {
		return StatusHealthy
	}

	hasDegraded := false

	for _, result := range results {
		status := normalizeStatus(result.Status)
		criticality := normalizeCriticality(result.Criticality)

		switch status {
		case StatusHealthy:
			continue
		case StatusDegraded:
			hasDegraded = true
		case StatusUnhealthy, StatusUnknown:
			if criticality == CriticalityOptional {
				hasDegraded = true
				continue
			}

			return StatusUnhealthy
		default:
			if criticality == CriticalityOptional {
				hasDegraded = true
				continue
			}

			return StatusUnhealthy
		}
	}

	if hasDegraded {
		return StatusDegraded
	}

	return StatusHealthy
}

func normalizeStatus(status Status) Status {
	switch status {
	case StatusHealthy, StatusDegraded, StatusUnhealthy, StatusUnknown:
		return status
	case "":
		return StatusUnknown
	default:
		return StatusUnknown
	}
}

func normalizeCriticality(criticality Criticality) Criticality {
	switch criticality {
	case CriticalityOptional:
		return CriticalityOptional
	default:
		return CriticalityRequired
	}
}
