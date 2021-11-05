// web based authentication
package middleware

import (
	"myapp/data"

	"github.com/cmd-ctrl-q/celeritas"
)

type Middleware struct {
	App    *celeritas.Celeritas
	Models data.Models
}
