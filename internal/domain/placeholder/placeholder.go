// Package placeholder decides when a placeholder flag clears (BR-2).
//
// Steven chose "auto-clear on edit" (question 14). Read literally that would
// mean ANY save clears the flag — including reordering a list or toggling
// Published — and a placeholder testimonial would quietly turn into a real
// one because someone moved it up. So the rule is narrower: the flag clears
// when a CONTENT field changes. Publishing, ordering and the flag itself are
// not content.
package placeholder

// Next returns the flag after a save.
//
//   - explicit: the editor set the checkbox in this same save. That always
//     wins (BR-2.3): an editor can mark a row placeholder again, or clear it
//     without changing the text.
//   - otherwise the flag clears if any content field changed (BR-2.2).
func Next(current bool, contentChanged bool, explicit *bool) bool {
	if explicit != nil {
		return *explicit
	}
	if current && contentChanged {
		return false
	}
	return current
}

// ContentChanged compares the content fields of a row before and after. A
// field present in `after` but absent in `before` counts as a change; values
// are compared by their string form, which is how they arrive from a form.
func ContentChanged(before, after map[string]string, contentFields []string) bool {
	for _, f := range contentFields {
		a, inAfter := after[f]
		if !inAfter {
			continue
		}
		if b, inBefore := before[f]; !inBefore || a != b {
			return true
		}
	}
	return false
}
