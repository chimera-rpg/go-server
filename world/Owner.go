package world

type Owner struct {
	commandQueue  []OwnerCommand
	repeatCommand OwnerCommand
	wizard        bool
}

func (owner *Owner) HasCommands() bool {
	return len(owner.commandQueue) > 0
}

func (owner *Owner) PushCommand(c OwnerCommand) {
	owner.commandQueue = append(owner.commandQueue, c)
}

func (owner *Owner) ShiftCommand() OwnerCommand {
	c := owner.commandQueue[0]
	owner.commandQueue = owner.commandQueue[1:]
	return c
}

func (owner *Owner) ClearCommands() {
	owner.commandQueue = make([]OwnerCommand, 0)
	owner.repeatCommand = nil
}

func (owner *Owner) RepeatCommand() OwnerCommand {
	return owner.repeatCommand
}

func (owner *Owner) Wizard() bool {
	return owner.wizard
}
