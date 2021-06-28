##### Reflex is a small tool to watch a directory and rerun a command when certain files change. It’s great for automatically running compile/lint/test tasks and for reloading your application when the code changes

This is exactly what we need to achieve our goal.

First we need to install reflex. `go get github.com/cespare/reflex`

We verify that the installation was successful with:

`server$ reflex`

`Could not make reflex for config: must give command to execute`

This means that the reflex command works, we just didn’t specify a valid command. Great!

The command we need is:

`reflex -r '\.go' -s -- sh -c "go run main.go"`

Let’s break down the different pieces:

`-r '.go'` Tells reflex to watch all files ending in go. If a change on such a file happens it will trigger the command

`-s` Tells reflex that we are running a service, as opposed to a command that will terminate by itself. If this flag is present, reflex will first kill the running service.

`sh -c "go run main.go"` the part after the — tells reflex which command to execute if a change was triggered.

With this in place, our go a server will now restart every time one of our go files was changed.