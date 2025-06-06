# Todo TUI Application Instructions

This TUI application helps you manage your tasks efficiently from the terminal. Below are the available modes and their controls:

## Global
- **s**: Switch between normal and completed modes. Pressing 's' toggles between your active (normal) tasks and completed tasks. (In the future, dedicated keys like 'c' for completed and 'n' for normal will be added.)

## Normal Mode
- View your list of tasks.
- The cursor points to the current task.
- **e**: Edit the selected task.
- **Up/Down Arrow**: Move the cursor to the previous/next task.
- **a**: Add a new task (switches to addition mode).
- **s**: Switch to completed mode.

## Addition Mode
- Enter the name of the new task.
- **Enter**: Add the task (if input is not empty) and return to normal mode. If input is empty, return to normal mode without adding.
- **Esc/q**: Cancel and return to normal mode.

## Completed Mode
- View your list of completed tasks.
- The cursor points to the current completed task.
- **Up/Down Arrow**: Move the cursor to the previous/next completed task.
- **s**: Switch to normal mode.

## Notes
- The application uses keyboard shortcuts for fast navigation and task management.
- Placeholders and prompts guide you during task addition and editing.
- All changes are reflected instantly in the TUI.