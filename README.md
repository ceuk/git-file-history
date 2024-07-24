# Git File History

Git File History is a lightweight, standalone tool that allows you to view all commits affecting a specific file.

![image](https://github.com/user-attachments/assets/5b5b890d-3da4-430c-a6d5-0dd94886b001)


## Features

- View commit history for a specific file.
- Page through commits to see changes.

## Installation

### Via the Go toolchain

```sh
go install github.com/ceuk/git-file-history@latest
npm i -g git-split-diffs
```

### Manually

Download the latest release from the [releases page](https://github.com/ceuk/git-file-history/releases).

### Arch (AUR)

Install the `git-file-history` package from the AUR.

## Usage

To use Git File History, simply run the command followed by the relative file path you want to inspect:

```sh
git-file-history <relative file path>
```

## Key Bindings

### List view

- `j`: Scroll down
- `k`: Scroll up
- `enter`: View the diff of the selected commit.
- `q`: Quit the application.

### Diff view

- `j`: Scroll diff down
- `k`: Scroll diff up
- `shift+k`: Go to the previous commit.
- `shift+j`: Go to the next commit.
- `g`: Go to the top of the diff.
- `G`: Go to the bottom of the diff.
- `q`: Go back to the list view.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on GitHub.
