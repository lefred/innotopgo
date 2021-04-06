package innotop

import (
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/text"
)

func DisplayHelp(c *container.Container) error {
	help_window, err := text.New()
	if err != nil {
		return err
	}
	c.Update("dyn_top_container", container.SplitHorizontal(container.Top(
		container.Border(linestyle.Light),
		container.ID("top_container"),
	),
		container.Bottom(
			container.Border(linestyle.Light),
			container.ID("main_container"),
			container.PlaceWidget(help_window),
			container.FocusedColor(cell.ColorNumber(15)),
		), container.SplitFixed(0)))
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("HELP (<-- <Backspace> to return to Processlist)"))
	help_window.Write("\n\n InnoTop Go Help\n")
	help_window.Write(" ===============\n\n")
	help_window.Write(" Main keys (available in all sections)       Help Screen (?)\n")
	help_window.Write(" -------------------------------------       ---------------\n\n")
	help_window.Write(" <ESC> : quit InnoTop any time               <backspace> : return to processlist\n")
	help_window.Write(" <?>   : get this screen\n\n")
	help_window.Write(" Processlist Screen                           Query Execution Plan Screen (E)\n")
	help_window.Write(" ------------------                           -------------------------------\n\n")
	help_window.Write(" <spacebar> : refresh processlist                        <backspace> : return to processlist\n")
	help_window.Write(" <E>        : go to Query Execution Plan                 <spacebar>  : change format of QEP\n")
	help_window.Write("                                                                       (normal, tree, json)\n")
	help_window.Write("                                              <mouse and arrow keys> : change the focus on section\n")
	help_window.Write("                                                                       and browse using the arrow keys\n")

	return nil
}
