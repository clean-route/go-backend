# CleanRoute Backend 
This is the backend repository of CleanRoute application. The parent repo can be found [here](https://github.com/sadityakumar9211/clean-route).

>> Please use bash/zsh shell for typing commands. Windows users please use GitBash for this purpose.

### IMPORTANT
Before you begin with frontend setup: Make sure you've completed the backend setup and the server is up and running.

Recommended steps for bootstrapping the servers: 
1. Setup and run the ML server
2. Setup and run Go backend server
3. Setup and run frontend client


### Prerequisites
1. Make sure you have Go installed in your system. This [article](https://mindmajix.com/how-to-install-golang) could help with that. 
 - You can verify that `Go` is installed by running `go version` command and getting a version similar to 1.20 or 1.21 or 1.22
2. Install `air` a tool similar to nodemon but for Go. Install through installation instruction from [this](https://mindmajix.com/how-to-install-golang) GitHub repository


### Installation instructions
1. Create a .env file and replace `xxxx` with your own secret values.
2. Just run `air` in terminal and see what happens, if it says `running...`, then you're good to go and your Go backend server is up and running. Otherwise, just try to debug the issue. Maybe you've not installed a tool properly, or it's not in your $PATH variable or something that I might have missed in the instructions, so please reach out to me to improve the instructions.
