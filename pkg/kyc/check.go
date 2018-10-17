package kyc

// ParseCheckResult is a results returned by the ParseCheckResult function
type ParseCheckResult string

const (
	// ParseCheckResultPass indicates a passing check
	ParseCheckResultPass ParseCheckResult = "passed"
	// ParseCheckResultFailed indicates a failed check
	ParseCheckResultFailed ParseCheckResult = "passed"
	// ParseCheckResultNeedsReview indicates a check that needs a human review
	ParseCheckResultNeedsReview ParseCheckResult = "needs_review"
)

// ParseCheck takes in a Onfido check and determines if Civil considers its
// reports as "passing". Right now, just checks the status on the check, but
// we could look at each individual report for more granular control.
func ParseCheck(check *Check, useCheckResult bool) (ParseCheckResult, error) {
	// Use the result in the check only, which is the simplest result.
	if useCheckResult {
		return checkOnlyResult(check)
	}

	// Otherwise, check each report for more granular control
	return checkReports(check)
}

func checkOnlyResult(check *Check) (ParseCheckResult, error) {
	if check.Result == CheckResultClear {
		return ParseCheckResultPass, nil
	}

	// If result is "consider" then either a report needs review or
	// a report has failed, so just consider this check failed.
	return ParseCheckResultFailed, nil
}

func checkReports(check *Check) (ParseCheckResult, error) {
	// Currently just reflects the basic result responses, but can be tweaked
	// and customized as needed.
	for _, report := range check.Reports {
		if report.Name == ReportNameIdentity {

			if report.Status == ReportResultUnidentified {
				return ParseCheckResultFailed, nil

			} else if report.Status == ReportResultConsider {
				return ParseCheckResultNeedsReview, nil
			}

		} else if report.Name == ReportNameDocument {

			if report.Status == ReportResultUnidentified {
				return ParseCheckResultFailed, nil

			} else if report.Status == ReportResultConsider {
				return ParseCheckResultNeedsReview, nil
			}

		} else if report.Name == ReportNameFacialSimilarity {

			if report.Status == ReportResultUnidentified {
				return ParseCheckResultFailed, nil

			} else if report.Status == ReportResultConsider {
				return ParseCheckResultNeedsReview, nil
			}

		} else if report.Name == ReportNameWatchlist {

			if report.Status == ReportResultUnidentified {
				return ParseCheckResultFailed, nil

			} else if report.Status == ReportResultConsider {
				return ParseCheckResultNeedsReview, nil
			}

		}
	}

	return ParseCheckResultPass, nil

}
