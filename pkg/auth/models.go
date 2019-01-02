package auth

import (
	"fmt"
	"io"
	"strconv"
)

// Token is a decoded JWT auth token
type Token struct {
	Sub     string
	IsAdmin bool
}

// LoginResponse is sent when a User successfully logs in
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	UID          string `json:"uid"`
}

// ApplicationEnum represents the different Civil applications supported
// by the auth system
type ApplicationEnum string

const (
	// ApplicationEnumDefault is the default application value
	ApplicationEnumDefault ApplicationEnum = "DEFAULT"
	// ApplicationEnumNewsroom is the newsroom signup application value
	ApplicationEnumNewsroom ApplicationEnum = "NEWSROOM"
	// ApplicationEnumStorefront is the storefront application value
	ApplicationEnumStorefront ApplicationEnum = "STOREFRONT"
)

// IsValid returns if the enum is a valid one
func (e ApplicationEnum) IsValid() bool {
	switch e {
	case ApplicationEnumDefault, ApplicationEnumNewsroom, ApplicationEnumStorefront:
		return true
	}
	return false
}

// String returns the string value of this enum
func (e ApplicationEnum) String() string {
	return string(e)
}

// UnmarshalGQL unmarshals from a value to the enum
func (e *ApplicationEnum) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApplicationEnum(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApplicationEnum", str)
	}
	return nil
}

// MarshalGQL marshals from an enum to the writer
func (e ApplicationEnum) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String())) // nolint: errcheck
}
