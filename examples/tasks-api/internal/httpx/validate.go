package httpx

// Validator collects field errors so a client sees every problem at once,
// not just the first one.
//
//	var v httpx.Validator
//	v.Check(req.Title != "", "title", "ห้ามว่าง")
//	if err := v.Err(); err != nil { return err }
type Validator struct {
	errs []FieldError
}

// Check records issue for field when ok is false.
func (v *Validator) Check(ok bool, field, issue string) {
	if !ok {
		v.errs = append(v.errs, FieldError{Field: field, Issue: issue})
	}
}

// Valid reports whether no checks have failed.
func (v *Validator) Valid() bool { return len(v.errs) == 0 }

// Err returns a 422 *APIError, or nil when every check passed.
func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}
	return Validation(v.errs)
}
