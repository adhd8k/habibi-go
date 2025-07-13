[] Creating a new project where a repo or the folder doesn't exist should 'just work', we should create directory if it doesn't exist, and initialize the repo, and create an initial commit with .gitignore having `.habibi-worktrees` 
[] Make sure the default worktrees location is .habibi-worktrees (overridable in config as is now, also make sure that override is working)
[] 'Restart Agent' upon restarting server or browser doesn't preseve message history, not sure what the solution is here, but maybe listing the agents within the session and having the user select one ( with the latest being on top /easily selectable ) so after restarting the server I can select a previous agent run and continue that instead of starting a new session
[] Limit git diff file size shown/returned - prevent huge git diffs from crashing the browser
[] Session created but still failed in the UI (refreshing ui shows the created session) - this is a very annoying bug as the network request is successful so must be a frontend issue
[] Terminal shouldn't restart  when switching tabs or sessions ( maybe initially just tabs if it's a lot easier, will need planning ) 
[] Update the assistant UI  - Remove tool calls from chat interface , just the interface, we will show them in the next tasks in different ways
[] Update the assistant UI  - By looking at ToDo related tool calls and results, Keep a ToDo list and update it according to the tool calls, should look nice
[] Update the assistant UI  - Show the latest tool call and result only on the side, below the todo list
[] Make command to build binaries for different OSes , let's start with MacOS, the binary should include the built web client as well
