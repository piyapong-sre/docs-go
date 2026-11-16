// ตัวอย่าง reflect: TypeOf/ValueOf, Kind, struct tag, CanSet และ validator อย่างง่าย
package main

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type SignupForm struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,contains=@"`
	Age   int    `json:"age" validate:"min=18"`
	Note  string `json:"note"` // ไม่มี validate tag -> ข้าม
}

// Validate ตรวจ field ตาม tag `validate` (รองรับ required, min=N, contains=X)
func Validate(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem() // รับได้ทั้ง struct และ *struct
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("validate: ต้องเป็น struct แต่ได้ %s", rv.Kind())
	}

	rt := rv.Type()
	var errs []error
	for i := range rt.NumField() {
		field := rt.Field(i) // reflect.StructField: ชื่อ, ชนิด, tag
		value := rv.Field(i) // reflect.Value: ค่าจริง
		tag, ok := field.Tag.Lookup("validate")
		if !ok {
			continue
		}
		for rule := range strings.SplitSeq(tag, ",") {
			name, arg, _ := strings.Cut(rule, "=")
			if err := check(field.Name, value, name, arg); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func check(field string, v reflect.Value, rule, arg string) error {
	switch rule {
	case "required":
		if v.IsZero() {
			return fmt.Errorf("%s: required", field)
		}
	case "min":
		n, _ := strconv.Atoi(arg)
		if v.Kind() == reflect.Int && v.Int() < int64(n) {
			return fmt.Errorf("%s: ต้องไม่น้อยกว่า %d", field, n)
		}
	case "contains":
		if v.Kind() == reflect.String && !strings.Contains(v.String(), arg) {
			return fmt.Errorf("%s: ต้องมี %q", field, arg)
		}
	}
	return nil
}

func main() {
	// --- TypeOf / ValueOf / Kind ---
	for _, x := range []any{42, "hi", []int{1}, map[string]int{}, SignupForm{}, &SignupForm{}} {
		t := reflect.TypeOf(x)
		fmt.Printf("%-18v kind=%s\n", t, t.Kind())
	}

	// --- อ่าน struct tag ---
	f, _ := reflect.TypeFor[SignupForm]().FieldByName("Email")
	fmt.Println("json tag:", f.Tag.Get("json"), "| validate tag:", f.Tag.Get("validate"))

	// --- CanSet: ต้องส่ง pointer และ field ต้อง exported ---
	form := SignupForm{Name: "Ann"}
	fmt.Println("CanSet (by value):", reflect.ValueOf(form).Field(0).CanSet())
	nameField := reflect.ValueOf(&form).Elem().FieldByName("Name")
	fmt.Println("CanSet (by pointer):", nameField.CanSet())
	nameField.SetString("Bob")
	fmt.Println("after SetString:", form.Name)

	// --- validator ---
	bad := SignupForm{Email: "bob.example.com", Age: 15}
	fmt.Println("--- validate bad ---")
	fmt.Println(Validate(bad))

	good := &SignupForm{Name: "Bob", Email: "bob@example.com", Age: 20}
	fmt.Println("--- validate good ---")
	fmt.Println(Validate(good))
}
