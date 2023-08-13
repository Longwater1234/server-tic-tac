# server-tic-tac

Server built specifically for this game [go-tic-tac](https://github.com/Longwater1234/go-tic-tac). You will only
need Golang 1.19 or higher installed. Download and install GO from [official link](https://go.dev/dl/), or use your OS
package manager.

## How to build

- Simply open up your terminal (or CMD) at this project root directory and run the following command.

    ```bash
    go build --ldflags="-s -w"
    ```
- Done. Now you can simply execute the binary `./server-tic-tac`.
- By default, it will run on port 9876.
- To stop the server anytime, just press `CTRL` + `C`, or quit the terminal window


## License 
&copy; 2023, Davis Tibbz, MIT License

## Contributions & Pull request

Pull requests and issues are welcome.