package style

import (
	"fmt"
	"strings"

	"github.com/jon4hz/deadshot/internal/ui/common"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var logo = fmt.Sprintf(`      _                _     _           _   
   __| | ___  __ _  __| |___| |__   ___ | |_ 
  / _P |/ _ \/ _P |/ _P / __| '_ \ / _ \| __|
 | (_| |  __/ (_| | (_| \__ \ | | | (_) | |_ 
  \__,_|\___|\__,_|\__,_|___/_| |_|\___/ \__|
                                              
  ( %s_<)︻┻┳══━一 - - - - - - - - - - - - 💥
`, termenv.String("φ").Foreground(common.Red.Color()).String())

// GenLogo returns the logo.
func GenLogo() string {
	if profile == termenv.Ascii {
		logo = strings.ReplaceAll(logo, "P", "`")
		return strings.ReplaceAll(logo, "💥", "X ")
	}

	x := strings.ReplaceAll(logo, "P", "`")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(x)
}
