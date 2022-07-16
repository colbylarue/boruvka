# Boruvka

Implementation of the boruvka MST algorithm. 
These instructions assume VS Code is being used. 
## Prerequisites
Install Git from 
Install Golang
Install Google Chrome (mine works in firefox and chrome)
Install Node.js from https://nodejs.org/en/download/
Install cesium from https://cesium.com/downloads/ using npm

```bash
npm install cesium 
```
## Setup
First clone git repo.
Navigate to graph directory in command line and install the local graph as a go package (Note: if you make changes to the graph library you may need to do a go install again)

```bash
cd wherever
git clone --recurse-submodules https://github.com/colbylarue/boruvka.git
cd boruvka
go install
```
Open VSCode. 
Open Folder. Select the folder you just cloned.

Install Visualization
```bash
cd cesium-webpack-example-main
npm install
npm start
```
## Usage

```
TODO:
```

## Contributing
Create a new branch to work in. 
Push your local changes to your remote branch
Create a Pull Request to merge into master. 
     a. select  cdlarue (me) as a reviewer.
     b. address review comments if applicable.
     b. merge when approved. 

## License
TODO