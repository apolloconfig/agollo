package perror

import "fmt"

var (
	ErrOverMaxRetryStill = fmt.Errorf("Over Max Retry Still Error")
	ErrUnauthorized      = fmt.Errorf("Unauthorized")
)
