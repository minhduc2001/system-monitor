package system

import (
    "os/exec"
    "runtime"
)

type Editor struct {
    Name string `json:"name"`
    Cmd  string `json:"cmd"`
}

func DetectEditors() []Editor {
    var editors []Editor
    os := runtime.GOOS

    if hasCommand("code") {
        editors = append(editors, Editor{"VSCode", "code"})
    }
    if hasCommand("idea") {
        editors = append(editors, Editor{"IntelliJ IDEA", "idea"})
    }
    if hasCommand("cursor") {
        editors = append(editors, Editor{"Cursor", "cursor"})
    }

    if os == "darwin" && !hasCommand("code") {
        editors = append(editors, Editor{"VSCode (App)", "open -a 'Visual Studio Code'"})
    }

    return editors
}

func hasCommand(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}
