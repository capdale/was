package email_test

import (
	"fmt"
	"testing"

	"github.com/capdale/was/email"
)

func TestCensorEmail(t *testing.T) {
	var cases = []struct {
		email string
		want  string
	}{
		{"testemail@domain.com", "t********@d*****.c**"},
		{"short@short.sho", "s****@s****.s**"},
		{"test@test.domain.com", "t***@t***.******.c**"},
	}

	for _, tcase := range cases {
		testname := fmt.Sprintf("%s, %s", tcase.email, tcase.want)
		t.Run(testname, func(t *testing.T) {
			censored := email.CensorEmail(tcase.email)
			if censored != tcase.want {
				t.Errorf("got %s, expected %s", censored, tcase.want)
			}
		})
	}
}
